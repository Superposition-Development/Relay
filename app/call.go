package app

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/gen2brain/malgo"
	"github.com/hraban/opus"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

const (
	sampleRate           = 48000
	audioChannels        = 2
	frameDurationMs      = 20
	frameSize            = sampleRate * frameDurationMs / 1000
	totalSamplesPerFrame = frameSize * audioChannels
	pcmBytesPerFrame     = totalSamplesPerFrame * 2
	maxAudioBufferSize   = 192000
)

var GlobalCallControl *CallControl

type MicStream struct {
	ctx      *malgo.AllocatedContext
	device   *malgo.Device
	outTrack *webrtc.TrackLocalStaticSample
	stopChan chan struct{}
	wg       sync.WaitGroup
}

type AudioBuffer struct {
	buf    []byte
	mu     sync.Mutex
	cond   *sync.Cond
	closed bool
}

type AudioEngine struct {
	otoCtx      *oto.Context
	player      *oto.Player
	audioReader *AudioBuffer
}

type CallControl struct {
	wsMu   sync.Mutex
	callID string

	pc          *webrtc.PeerConnection
	audioEngine *AudioEngine
	micStream   *MicStream
	outTrack    *webrtc.TrackLocalStaticSample

	pendingCandidates []webrtc.ICECandidateInit
	remoteDescSet     bool

	mu sync.Mutex
}

func NewMicStream(outTrack *webrtc.TrackLocalStaticSample) (*MicStream, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init malgo context: %w", err)
	}

	return &MicStream{
		ctx:      ctx,
		outTrack: outTrack,
		stopChan: make(chan struct{}),
	}, nil
}

func (m *MicStream) Start() error {
	encoder, err := opus.NewEncoder(sampleRate, audioChannels, opus.AppVoIP)
	if err != nil {
		return fmt.Errorf("failed to create opus encoder: %w", err)
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = audioChannels
	deviceConfig.SampleRate = sampleRate
	deviceConfig.Alsa.NoMMap = 1

	pcmChan := make(chan []byte, 64)

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: func(_, pInputSamples []byte, _ uint32) {
			if len(pInputSamples) == 0 {
				return
			}
			buf := make([]byte, len(pInputSamples))
			copy(buf, pInputSamples)

			select {
			case pcmChan <- buf:
			default:
			}
		},
	}

	device, err := malgo.InitDevice(m.ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return fmt.Errorf("failed to init capture device: %w", err)
	}

	if err := device.Start(); err != nil {
		device.Uninit()
		return fmt.Errorf("failed to start capture device: %w", err)
	}
	m.device = device

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		var pcmAccumulator []byte
		int16Buf := make([]int16, totalSamplesPerFrame)
		opusBuf := make([]byte, 4000)

		for {
			select {
			case <-m.stopChan:
				return
			case chunk := <-pcmChan:
				pcmAccumulator = append(pcmAccumulator, chunk...)

				for len(pcmAccumulator) >= pcmBytesPerFrame {
					frameBytes := pcmAccumulator[:pcmBytesPerFrame]
					pcmAccumulator = pcmAccumulator[pcmBytesPerFrame:]

					for i := 0; i < totalSamplesPerFrame; i++ {
						int16Buf[i] = int16(binary.LittleEndian.Uint16(frameBytes[i*2 : i*2+2]))
					}

					n, err := encoder.Encode(int16Buf, opusBuf)
					if err != nil || n == 0 {
						continue
					}

					sampleData := make([]byte, n)
					copy(sampleData, opusBuf[:n])

					if m.outTrack != nil {
						if err := m.outTrack.WriteSample(media.Sample{
							Data:     sampleData,
							Duration: frameDurationMs * time.Millisecond,
						}); err != nil {

						}
					}
				}
			}
		}
	}()

	return nil
}

func InstantiateCallControl(callID any) *CallControl {
	return &CallControl{
		callID: fmt.Sprintf("%v", callID),
	}
}

func (m *MicStream) Stop() {
	select {
	case <-m.stopChan:
		return
	default:
		close(m.stopChan)
	}

	if m.device != nil {
		_ = m.device.Stop()
		m.device.Uninit()
		m.device = nil
	}

	m.wg.Wait()

	if m.ctx != nil {
		m.ctx.Free()
		m.ctx = nil
	}
}

func NewAudioBuffer() *AudioBuffer {
	b := &AudioBuffer{
		buf: make([]byte, 0, maxAudioBufferSize),
	}
	b.cond = sync.NewCond(&b.mu)
	return b
}

func (b *AudioBuffer) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for len(b.buf) == 0 && !b.closed {
		b.cond.Wait()
	}

	if b.closed && len(b.buf) == 0 {
		return 0, io.EOF
	}

	n := copy(p, b.buf)
	b.buf = b.buf[n:]
	return n, nil
}

func (b *AudioBuffer) Write(p []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.buf = append(b.buf, p...)
	if len(b.buf) > maxAudioBufferSize {
		excess := len(b.buf) - maxAudioBufferSize
		b.buf = b.buf[excess:]
	}

	b.cond.Signal()
}

func (b *AudioBuffer) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
	b.cond.Broadcast()
}

func NewAudioEngine() (*AudioEngine, error) {
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: audioChannels,
		Format:       oto.FormatSignedInt16LE,
	}

	otoCtx, ready, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio context: %w", err)
	}

	<-ready

	audioReader := NewAudioBuffer()
	player := otoCtx.NewPlayer(audioReader)
	go player.Play()

	return &AudioEngine{
		otoCtx:      otoCtx,
		player:      player,
		audioReader: audioReader,
	}, nil
}

func (a *AudioEngine) HandleRemoteTrack(track *webrtc.TrackRemote) {
	decoder, err := opus.NewDecoder(sampleRate, audioChannels)
	if err != nil {
		return
	}

	pcmInt16Buf := make([]int16, 5760*audioChannels)
	pcmByteBuf := make([]byte, len(pcmInt16Buf)*2)

	for {
		rtpPacket, _, readErr := track.ReadRTP()
		if readErr != nil {
			if readErr != io.EOF {
			}
			return
		}

		if len(rtpPacket.Payload) == 0 {
			continue
		}

		samplesDecoded, err := decoder.Decode(rtpPacket.Payload, pcmInt16Buf)
		if err != nil || samplesDecoded == 0 {
			continue
		}

		totalSamples := samplesDecoded * audioChannels
		for i := 0; i < totalSamples; i++ {
			binary.LittleEndian.PutUint16(pcmByteBuf[i*2:i*2+2], uint16(pcmInt16Buf[i]))
		}

		if a.audioReader != nil {
			a.audioReader.Write(pcmByteBuf[:totalSamples*2])
		}
	}
}

func (a *AudioEngine) Close() {
	if a.audioReader != nil {
		a.audioReader.Close()
	}

	if a.player != nil {
		a.player = nil
	}

	if a.otoCtx != nil {
		_ = a.otoCtx.Suspend()
	}
}

var ClientPeerConnection *webrtc.PeerConnection

// var config = webrtc.Configuration{
// 	ICEServers: []webrtc.ICEServer{
// 		{
// 			URLs: []string{
// 				"turn:global.relay.metered.ca:80",
// 				"turn:global.relay.metered.ca:443",
// 				"turn:global.relay.metered.ca:443?transport=tcp",
// 			},
// 			Username:   "a072cb146b471d7876e641dc",
// 			Credential: "AbV/kjuHbgOurcxl",
// 		},
// 	},
// 	ICETransportPolicy: webrtc.ICETransportPolicyRelay,
// }

var config = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{URLs: []string{"stun:stun.l.google.com:19302"}},
	},
}

// var config = webrtc.Configuration{
// 	ICEServers: []webrtc.ICEServer{
// 		{
// 			URLs: []string{
// 				// "turn:global.relay.metered.ca:80",
// 				// "turn:global.relay.metered.ca:443",
// 				"turn:global.relay.metered.ca:443?transport=tcp",
// 			},
// 			Username:   "c1bea89d980d944a146c66a3",
// 			Credential: "RbBZdljmQEoTFBC+",
// 		},
// 	},
// 	// ICETransportPolicy: webrtc.ICETransportPolicyRelay,
// }

func (m *CallControl) JoinCall() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var err error
	m.audioEngine, err = NewAudioEngine()
	if err != nil {
		panic(err)
	}

	token, err := LoadToken()
	cpc, err := webrtc.NewPeerConnection(config)
	ClientPeerConnection = cpc
	m.pc = ClientPeerConnection
	if err != nil {
		panic(err)
	}

	ClientPeerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		SendWebsocketJSON(map[string]interface{}{
			"candidate": c.ToJSON(),
			"callID":    "dietz",
			"type":      "candidate",
			"authKey":   token,
		})
	})

	ClientPeerConnection.OnTrack(func(track *webrtc.TrackRemote, receiever *webrtc.RTPReceiver) {
		m.mu.Lock()
		engine := m.audioEngine
		m.mu.Unlock()

		if engine != nil {
			go engine.HandleRemoteTrack(track)
		}
	})

	// localDescription, err := ClientPeerConnection.CreateOffer(&webrtc.OfferOptions{
	// 	OfferAnswerOptions: webrtc.OfferAnswerOptions{},
	// 	ICERestart:         true,
	// })

	uniqueTrackID := fmt.Sprintf("%s-%s", CurrentUserID, "dietz")
	streamID := fmt.Sprintf("stream-%s", CurrentUserID)

	m.outTrack, err = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		uniqueTrackID,
		streamID,
	)

	if err != nil {
		return
	}

	m.micStream, err = NewMicStream(m.outTrack)
	if err == nil {
		_ = m.micStream.Start()
	}

	if _, err := m.pc.AddTrack(m.outTrack); err != nil {
		return
	}

	localDescription, err := ClientPeerConnection.CreateOffer(nil)

	if err != nil {
		panic(err)
	}

	ClientPeerConnection.SetLocalDescription(localDescription)

	payload := map[string]any{
		"authKey": token,
		"callID":  "dietz",
		"offer":   localDescription,
		"type":    "joinCall",
	}
	SendWebsocketJSON(payload)
}

func LeaveCall() {
	token, err := LoadToken()
	if err != nil {
		//oh well ig
	}
	payload := map[string]any{
		"callID":  "dietz",
		"authKey": token,
		"type":    "leaveCall",
	}
	SendWebsocketJSON(payload)
}
