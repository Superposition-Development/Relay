package app

import (
	"Relay/util"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var (
	mu sync.Mutex
)

type GetChannelsRequest struct {
	ServerID any `json:"serverID"`
}

func GetServers() []Server {
	JWTCookie, err := LoadToken()
	if err != nil {
		fmt.Print(err)
	}
	loginURL := url.URL{
		Scheme: ServerURL.Scheme,
		Host:   ServerURL.Host,
		Path:   GetServerEndpoint,
	}
	res := GET(JWTCookie, loginURL.String())
	if res.Error != nil {
		fmt.Println("Error:", res.Error)
		return nil
	}

	var realServers []Server
	if err := json.Unmarshal(res.Data, &realServers); err != nil {
		fmt.Println("Error decoding servers:", err)
		return nil
	}

	createServerItem := Server{
		ID:   "Six Seven",
		Name: "+ Create",
	}

	joinServerItem := Server{
		ID:   "Six Seven",
		Name: "+ Join",
	}

	selectDMItem := Server{
		ID:   "Six Seven",
		Name: "+ DM",
	}

	servers := append([]Server{createServerItem, joinServerItem, selectDMItem}, realServers...)

	ServerListToDataMap = make(map[int]Server, len(servers))
	for i, s := range servers {
		ServerListToDataMap[i] = s
	}

	return servers
}

func GetDMs() []DM {
	JWTCookie, err := LoadToken()
	if err != nil {
		fmt.Print(err)
	}
	GetDMURL := url.URL{
		Scheme: ServerURL.Scheme,
		Host:   ServerURL.Host,
		Path:   GetDMEndpoint,
	}
	res := GET(JWTCookie, GetDMURL.String())
	if res.Error != nil {
		fmt.Println("Error:", res.Error)
		return nil
	}

	var realDMs map[string][]string //[]DM
	if err := json.Unmarshal(res.Data, &realDMs); err != nil {
		fmt.Println("Error decoding servers:", err)
		return nil
	}

	createDMItem := DM{
		ID:     "Six Seven",
		UserID: "+ Create",
	}

	dms := []DM{createDMItem}

	for key, value := range realDMs {
		value = util.RemoveByValue(value, CurrentUserID)
		// fmt.Printf("Key: %d, Value: %s\n", key, value)
		dms = append(dms, DM{ID: key, UserID: value[0]})
	}

	DMListToDataMap = make(map[int]DM, len(dms))
	for i, s := range dms {
		DMListToDataMap[i] = s
	}

	return dms
	// return nil
}

func GetChannels(serverID any) []Channel {
	JWTCookie, err := LoadToken()
	if err != nil {
		fmt.Print(err)
	}
	reqPayload := GetChannelsRequest{
		ServerID: fmt.Sprintf("%v", serverID),
	}
	getChannelURL := url.URL{
		Scheme: ServerURL.Scheme,
		Host:   ServerURL.Host,
		Path:   GetChannelEndpoint,
	}

	var realChannels []Channel
	if err := POST(reqPayload, JWTCookie, getChannelURL.String(), &realChannels); err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	createChannelItem := Channel{
		ID:   "Six Seven",
		Name: "Create",
		Type: "Create",
	}

	channels := append([]Channel{createChannelItem}, realChannels...)

	ChannelListToDataMap = make(map[int]Channel, len(channels))
	for i, s := range channels {
		ChannelListToDataMap[i] = s
	}

	return channels
}

// return all the userIDs, this eventually will need to be updated for pagination purposes
func GetServerUsers(serverID any) []string {
	JWTCookie, err := LoadToken()
	if err != nil {
		fmt.Print(err)
	}
	reqPayload := GetChannelsRequest{ //we can steal this because i dont want to define a new type and this one matches the args
		ServerID: fmt.Sprintf("%v", serverID),
	}
	getChannelURL := url.URL{
		Scheme: ServerURL.Scheme,
		Host:   ServerURL.Host,
		Path:   GetServerUsersEndpoint,
	}

	var userIDs []string
	if err := POST(reqPayload, JWTCookie, getChannelURL.String(), &userIDs); err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return userIDs
}

type ResponseResult struct {
	Data  []byte
	Error error
}

func GET(authKey string, address string) ResponseResult {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return ResponseResult{Data: nil, Error: err}
	}

	req.Header.Set("Content-Type", "application/json")
	if authKey != "" {
		req.Header.Set("Authorization", "Bearer "+authKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return ResponseResult{Data: nil, Error: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ResponseResult{
			Data:  nil,
			Error: fmt.Errorf("network response failed with status: %s", resp.Status),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseResult{Data: nil, Error: err}
	}

	return ResponseResult{Data: body, Error: nil}
}

// directly encodes the destination result
func POST(payload any, authKey string, address string, response any) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		address,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authKey)

	req.Header.Set("Cookie", "RelayJWT="+authKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("network response failed [%s]: %s", resp.Status, string(body))
	}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}

	return nil
}

func GetConn() *websocket.Conn {
	mu.Lock()
	defer mu.Unlock()
	return Socket
}

func RegisterWebsocket(address string, jwtToken string) {
	formattedAddr := address

	if !strings.HasPrefix(formattedAddr, "ws://") &&
		!strings.HasPrefix(formattedAddr, "wss://") {
		formattedAddr = "ws://" + formattedAddr
	}

	conn, resp, err := websocket.DefaultDialer.Dial(formattedAddr, nil)
	if err != nil {
		if resp != nil {
		} else {
		}
		return
	}

	mu.Lock()
	Socket = conn
	mu.Unlock()

	registerMessage := map[string]any{
		"message": "register",
		"authKey": jwtToken,
	}

	SendWebsocketJSON(registerMessage)

	go func() {
		defer func() {
			mu.Lock()

			if Socket == conn {
				Socket = nil
			}

			mu.Unlock()

			_ = conn.Close()

		}()

		for {
			_, messageData, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("[WS] Read error: %v\n", err)
				return
			}

			var parsed map[string]any

			if err := json.Unmarshal(messageData, &parsed); err != nil {
				fmt.Printf("[WS] JSON decode error: %v\n", err)
				continue
			}
			WSChan <- parsed
			ObtainEvent(WebsocketMessage{Type: fmt.Sprintf("%v", parsed["type"]),
				Data: parsed["data"]})

		}
	}()
}

func SendWebsocketJSON(message any) {
	mu.Lock()
	defer mu.Unlock()

	if Socket == nil {
		return
	}

	err := Socket.WriteJSON(message)
	if err != nil {

	}
}

var WSChan = make(chan map[string]any, 100)

type WebsocketMessage struct {
	Type string `json:"type"`

	Data interface{} `json:"data"`
}

func ListenForWSMsg() tea.Cmd {
	return func() tea.Msg {
		msg := <-WSChan

		msgType, ok := msg["message"].(string)
		if !ok {
			// fmt.Println(msg)
			return nil
		}

		return WebsocketMessage{
			Type: msgType,
			Data: msg["data"],
		}
	}
}

func ObtainEvent(message WebsocketMessage) {
	switch message.Type {
	case "answer":
		var answerSDP string
		switch v := message.Data.(type) {
		case string:
			var answerMap map[string]interface{}
			if err := json.Unmarshal([]byte(v), &answerMap); err != nil {
				return
			}
			answerSDP, _ = answerMap["sdp"].(string)
		case map[string]interface{}:
			answerSDP, _ = v["sdp"].(string)
		}

		if answerSDP == "" {
			return
		}

		answer := webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  answerSDP,
		}

		if err := ClientPeerConnection.SetRemoteDescription(answer); err != nil {
			return
		}

	case "candidate":
		var init webrtc.ICECandidateInit //ICECandidateInit
		switch v := message.Data.(type) {
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
		ClientPeerConnection.AddICECandidate(init)
		//refer to repo prending candidates
	}

}
