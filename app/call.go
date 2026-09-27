package app

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type CallControl struct {
	wsMu   sync.Mutex
	callID string

	// pc *webrtc.PeerConnection
	// audioEngine *AudioEngine
	// micStream   *MicStream
	// outTrack    *webrtc.TrackLocalStaticSample

	pendingCandidates []webrtc.ICECandidateInit
	remoteDescSet     bool

	mu sync.Mutex
}

var ClientPeerConnection *webrtc.PeerConnection

var config = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
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
	ICETransportPolicy: webrtc.ICETransportPolicyRelay,
}

func StartCall() {
	cpc, err := webrtc.NewPeerConnection(config)
	ClientPeerConnection = cpc
	if err != nil {

	}

	localDescription, err := ClientPeerConnection.CreateOffer(&webrtc.OfferOptions{
		OfferAnswerOptions: webrtc.OfferAnswerOptions{},
		ICERestart:         true,
	})

	ClientPeerConnection.SetLocalDescription(localDescription)

	ClientPeerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		SendWebsocketJSON(map[string]interface{}{
			"candidate": c.ToJSON(),
			"message":   "candidate"})
	})

	token, err := LoadToken()
	payload := map[string]any{
		"callID":  "dietz", //fmt.Sprintf("%v", app.ChannelListToDataMap[m.activeChannelIndex].ID),
		"offfer":  localDescription,
		"authKey": token,
		"message": "joinCall",
	}
	SendWebsocketJSON(payload)

}
