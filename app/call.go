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

func JoinCall() {
	token, err := LoadToken()
	cpc, err := webrtc.NewPeerConnection(config)
	ClientPeerConnection = cpc
	if err != nil {

	}

	ClientPeerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		SendWebsocketJSON(map[string]interface{}{
			"candidate": c.ToJSON(),
			"message":   "candidate",
			"authKey":   token,
		})
	})

	// localDescription, err := ClientPeerConnection.CreateOffer(&webrtc.OfferOptions{
	// 	OfferAnswerOptions: webrtc.OfferAnswerOptions{},
	// 	ICERestart:         true,
	// })

	localDescription, err := ClientPeerConnection.CreateOffer(nil)

	if err != nil {
		panic(err)
	}

	// fmt.Println(localDescription)
	// currentServerAddress = localdescription

	ClientPeerConnection.SetLocalDescription(localDescription)
	payload := map[string]any{
		"authKey": token,
		"callID":  "dietz", //fmt.Sprintf("%v", app.ChannelListToDataMap[m.activeChannelIndex].ID),
		"offer":   localDescription,
		"message": "joinCall",
	}
	SendWebsocketJSON(payload)
}

func LeaveCall() {
	token, err := LoadToken()
	if err != nil {
		//oh well ig
	}
	payload := map[string]any{
		"callID":  "dietz", //fmt.Sprintf("%v", app.ChannelListToDataMap[m.activeChannelIndex].ID),
		"authKey": token,
		"message": "leaveCall",
	}
	SendWebsocketJSON(payload)
}
