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
	"github.com/gorilla/websocket"
	"github.com/hraban/opus"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type MicStream struct {
	ctx      *malgo.AllocatedContext
	device   *malgo.Device
	outTrack *webrtc.TrackLocalStaticSample
	stopChan chan struct{}
	wg       sync.WaitGroup
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

	const sampleRate = 48000
	const channels = 2
	const frameDurationMs = 20

	const frameSize = sampleRate * frameDurationMs / 1000
	const totalSamplesPerFrame = frameSize * channels
	const pcmBytesPerFrame = totalSamplesPerFrame * 2

	encoder, err := opus.NewEncoder(
		sampleRate,
		channels,
		opus.Application(opus.AppVoIP),
	)
	if err != nil {
		return fmt.Errorf("failed to create opus encoder: %w", err)
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = channels
	deviceConfig.SampleRate = sampleRate
	deviceConfig.Alsa.NoMMap = 1

	pcmChan := make(chan []byte, 64)

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: func(pOutputSamples, pInputSamples []byte, frameCount uint32) {

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

	device, err := malgo.InitDevice(
		m.ctx.Context,
		deviceConfig,
		deviceCallbacks,
	)
	if err != nil {
		return fmt.Errorf("failed to init capture device: %w", err)
	}

	if err := device.Start(); err != nil {
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
						int16Buf[i] = int16(
							binary.LittleEndian.Uint16(
								frameBytes[i*2 : i*2+2],
							),
						)
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
							Duration: time.Duration(frameDurationMs) * time.Millisecond,
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
		buf: make([]byte, 0, 192000),
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

	const maxBufferSize = 192000

	if len(b.buf) > maxBufferSize {
		excess := len(b.buf) - maxBufferSize
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

	const sampleRate = 48000
	const channelCount = 2
	const format = oto.FormatSignedInt16LE

	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		Format:       format,
	}

	otoCtx, ready, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio context: %w", err)
	}

	<-ready

	audioReader := NewAudioBuffer()

	player := otoCtx.NewPlayer(audioReader)

	go func() {
		player.Play()
	}()

	return &AudioEngine{
		otoCtx:      otoCtx,
		player:      player,
		audioReader: audioReader,
	}, nil
}

func (a *AudioEngine) HandleRemoteTrack(track *webrtc.TrackRemote) {

	const sampleRate = 48000
	const channels = 2

	decoder, err := opus.NewDecoder(sampleRate, channels)
	if err != nil {
		return
	}

	pcmInt16Buf := make([]int16, 5760*channels)
	pcmByteBuf := make([]byte, len(pcmInt16Buf)*2)

	rtpBuf := make([]byte, 2000)

	packetCount := 0

	for {
		n, _, err := track.Read(rtpBuf)
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}

		packet := &rtp.Packet{}
		if err := packet.Unmarshal(rtpBuf[:n]); err != nil {
			continue
		}

		packetCount++
		if packetCount <= 10 {
			fmt.Printf("RTP packet %d: payload=%d bytes\n", packetCount, len(packet.Payload))
		}

		if len(packet.Payload) == 0 {
			continue
		}

		samplesDecoded, err := decoder.Decode(packet.Payload, pcmInt16Buf)
		if err != nil {
			continue
		}

		if samplesDecoded == 0 {
			continue
		}

		totalSamples := samplesDecoded * channels
		for i := 0; i < totalSamples; i++ {
			binary.LittleEndian.PutUint16(
				pcmByteBuf[i*2:i*2+2],
				uint16(pcmInt16Buf[i]),
			)
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

	if a.otoCtx != nil {
		_ = a.otoCtx.Suspend()
	}
}

type CallControl struct {
	wsMu *sync.Mutex
	ws   *websocket.Conn

	pc *webrtc.PeerConnection

	audioEngine       *AudioEngine
	micStream         *MicStream
	outTrack          *webrtc.TrackLocalStaticSample
	pendingCandidates []webrtc.ICECandidateInit
	remoteDescSet     bool
}

func InstantiateCallControl() *CallControl {

	return &CallControl{
		wsMu: &mu,
		ws:   GetConn(),
	}
}

func (m *CallControl) sendWSMessage(Type string, Content map[string]interface{}) error {

	authKey, err := LoadToken()

	payload := map[string]interface{}{
		"authKey": authKey,
		"message": Type,
		"userID":  "poop",
		"callID":  "fortnite",
	}

	for k, v := range Content {
		payload[k] = v
	}

	SendWebsocketJSON(payload)
	return err
}

func (m *CallControl) StartCallRoutine() tea.Cmd {
	return func() tea.Msg {

		var err error

		if m.ws == nil {
			fmt.Println("no websocket")
			return nil
		}

		m.audioEngine, err = NewAudioEngine()
		if err != nil {
			fmt.Println(err)
			return nil
		}

		config := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{URLs: []string{"stun:stun.l.google.com:19302"}},
			},
		}

		m.pc, err = webrtc.NewPeerConnection(config)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		m.pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {

			if m.audioEngine != nil {
				go m.audioEngine.HandleRemoteTrack(track)
			}
		})

		m.pc.OnICECandidate(func(c *webrtc.ICECandidate) {

			if c == nil {
				return
			}

			_ = m.sendWSMessage("candidate", map[string]interface{}{
				"candidate": c.ToJSON(),
			})
		})

		m.outTrack, err = webrtc.NewTrackLocalStaticSample(
			webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
			"audio",
			"pion",
		)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if _, err = m.pc.AddTrack(m.outTrack); err != nil {
			fmt.Println(err)
			return nil
		}

		m.micStream, err = NewMicStream(m.outTrack)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if err := m.micStream.Start(); err != nil {
			fmt.Println(err)
			return nil
		}

		_ = m.sendWSMessage("joinCall", nil)

		return nil
	}
}

func (m *CallControl) HandleSignalMessage(msg WebsocketMesssage) {
	switch msg.Type {

	case "offer":
		var offerSDP string
		switch v := msg.Data.(type) {
		case string:
			var offerMap map[string]interface{}
			if err := json.Unmarshal([]byte(v), &offerMap); err == nil {
				offerSDP, _ = offerMap["sdp"].(string)
			}
		case map[string]interface{}:
			offerSDP, _ = v["sdp"].(string)
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

		for _, cand := range m.pendingCandidates {
			_ = m.pc.AddICECandidate(cand)
		}
		m.pendingCandidates = nil

		answer, err := m.pc.CreateAnswer(nil)
		if err != nil {
			return
		}

		if err := m.pc.SetLocalDescription(answer); err != nil {
			return
		}

		_ = m.sendWSMessage("answer", map[string]interface{}{
			"answer": map[string]interface{}{
				"type": "answer",
				"sdp":  answer.SDP,
			},
		})

	case "candidate":
		if candStr, ok := msg.Data.(string); ok && candStr != "" {
			var init webrtc.ICECandidateInit
			if err := json.Unmarshal([]byte(candStr), &init); err != nil {
				return
			}

			if !m.remoteDescSet {
				m.pendingCandidates = append(m.pendingCandidates, init)
				return
			}

			_ = m.pc.AddICECandidate(init)
		}
	}
}

func (m *CallControl) Close() {

	if m.micStream != nil {
		m.micStream.Stop()
		m.micStream = nil
	}

	if m.audioEngine != nil {
		m.audioEngine.Close()
		m.audioEngine = nil
	}

	if m.pc != nil {
		_ = m.pc.Close()
		m.pc = nil
	}

	_ = m.sendWSMessage("leaveCall", nil)
}
