package app

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ebitengine/oto/v3"
	"github.com/gen2brain/malgo"
	"github.com/hraban/opus"
	"github.com/pion/rtp"
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

type MicStream struct {
	ctx      *malgo.AllocatedContext
	device   *malgo.Device
	outTrack *webrtc.TrackLocalStaticSample
	stopChan chan struct{}
	wg       sync.WaitGroup
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
		opusBuf := make([]byte, 1000)

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
						_ = m.outTrack.WriteSample(media.Sample{
							Data:     sampleData,
							Duration: frameDurationMs * time.Millisecond,
						})
					}
				}
			}
		}
	}()

	return nil
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
		_ = m.ctx.Uninit()
		m.ctx.Free()
		m.ctx = nil
	}
}

type AudioBuffer struct {
	buf    []byte
	mu     sync.Mutex
	cond   *sync.Cond
	closed bool
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

type AudioEngine struct {
	otoCtx      *oto.Context
	player      *oto.Player
	audioReader *AudioBuffer
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
	rtpBuf := make([]byte, 2000)
	packetCount := 0

	for {
		n, _, err := track.Read(rtpBuf)
		if err != nil {
			if err != io.EOF {

			}
			return
		}

		packet := &rtp.Packet{}
		if err := packet.Unmarshal(rtpBuf[:n]); err != nil || len(packet.Payload) == 0 {
			continue
		}

		packetCount++
		if packetCount <= 10 {

		}

		samplesDecoded, err := decoder.Decode(packet.Payload, pcmInt16Buf)
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

func InstantiateCallControl(callID any) *CallControl {
	return &CallControl{
		callID: fmt.Sprintf("%v", callID),
	}
}

func (m *CallControl) sendWSMessage(messageType string, content map[string]interface{}) error {
	m.wsMu.Lock()
	defer m.wsMu.Unlock()

	authKey, err := LoadToken()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"authKey": authKey,
		"message": messageType,
		"callID":  m.callID,
	}

	for k, v := range content {
		payload[k] = v
	}

	SendWebsocketJSON(payload)
	return nil
}

func (m *CallControl) StartCallRoutine() tea.Cmd {
	return func() tea.Msg {
		if GetConn() == nil {

			return nil
		}

		m.mu.Lock()
		defer m.mu.Unlock()

		var err error
		m.audioEngine, err = NewAudioEngine()
		if err != nil {

			return nil
		}

		config := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
				{
					URLs: []string{
						"turn:global.relay.metered.ca:80",
						"turn:global.relay.metered.ca:443",
						"turn:global.relay.metered.ca:443?transport=tcp",
					},
					Username:   "a072cb146b471d7876e641dc",
					Credential: "AbV/kjuHbgOurcxl",
				},
			},
		}

		m.pc, err = webrtc.NewPeerConnection(config)
		if err != nil {

			return nil
		}

		// m.pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {

		// })

		// m.pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {

		// })

		m.pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			m.mu.Lock()
			engine := m.audioEngine
			m.mu.Unlock()

			if engine != nil {
				go engine.HandleRemoteTrack(track)
			}
		})

		m.pc.OnICECandidate(func(c *webrtc.ICECandidate) {
			if c == nil {

				return
			}

			if err := m.sendWSMessage("candidate", map[string]interface{}{
				"candidate": c.ToJSON(),
			}); err != nil {

			}
		})

		m.outTrack, err = webrtc.NewTrackLocalStaticSample(
			webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
			"audio",
			"pion",
		)
		if err != nil {

			return nil
		}

		if _, err := m.pc.AddTrack(m.outTrack); err != nil {

			return nil
		}

		m.micStream, err = NewMicStream(m.outTrack)
		if err != nil {

			return nil
		}

		if err := m.micStream.Start(); err != nil {

			return nil
		}

		if err := m.sendWSMessage("joinCall", nil); err != nil {

		}

		return nil
	}
}

func (m *CallControl) HandleSignalMessage(msg WebsocketMesssage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pc == nil {
		return
	}

	switch msg.Type {
	case "offer":
		m.handleOffer(msg)
	case "candidate":
		m.handleCandidate(msg)
	case "answer":

	}
}

func (m *CallControl) handleOffer(msg WebsocketMesssage) {
	var offerSDP string

	switch v := msg.Data.(type) {
	case string:
		var offerMap map[string]interface{}
		if err := json.Unmarshal([]byte(v), &offerMap); err != nil {
			return
		}
		offerSDP, _ = offerMap["sdp"].(string)
	case map[string]interface{}:
		offerSDP, _ = v["sdp"].(string)
	default:
		return
	}

	if offerSDP == "" {
		return
	}

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}

	if err := m.pc.SetRemoteDescription(offer); err != nil {
		return
	}

	m.remoteDescSet = true

	for _, candidate := range m.pendingCandidates {
		if err := m.pc.AddICECandidate(candidate); err != nil {

		}
	}
	m.pendingCandidates = nil

	answer, err := m.pc.CreateAnswer(nil)
	if err != nil {

		return
	}

	if err := m.pc.SetLocalDescription(answer); err != nil {

		return
	}

	if err := m.sendWSMessage("answer", map[string]interface{}{
		"answer": map[string]interface{}{
			"type": "answer",
			"sdp":  answer.SDP,
		},
	}); err != nil {

	}
}

func (m *CallControl) handleCandidate(msg WebsocketMesssage) {
	var init webrtc.ICECandidateInit

	switch v := msg.Data.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &init); err != nil {
			return
		}
	case map[string]interface{}:
		raw, err := json.Marshal(v)
		if err != nil {
			return
		}
		if err := json.Unmarshal(raw, &init); err != nil {
			return
		}
	default:
		return
	}

	if !m.remoteDescSet {
		m.pendingCandidates = append(m.pendingCandidates, init)
		return
	}

	if err := m.pc.AddICECandidate(init); err != nil {

	}
}

func (m *CallControl) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.micStream != nil {
		m.micStream.Stop()
		m.micStream = nil
	}

	if m.audioEngine != nil {
		aEngine := m.audioEngine
		m.audioEngine = nil
		aEngine.Close()
	}

	if m.pc != nil {
		_ = m.pc.Close()
		m.pc = nil
	}

	_ = m.sendWSMessage("leaveCall", nil)
}
