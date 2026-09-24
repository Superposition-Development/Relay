package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

var (
	mu sync.Mutex
)

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

			// fmt.Printf("[WS] RECEIVED RAW: %s\n", string(messageData))

			var parsed map[string]any

			if err := json.Unmarshal(messageData, &parsed); err != nil {
				fmt.Printf("[WS] JSON decode error: %v\n", err)
				continue
			}

			// fmt.Printf("[WS] RECEIVED TYPE: %v\n", parsed["message"])

			// if GlobalCallControl != nil {
			// 	GlobalCallControl.HandleSignalMessage(WebsocketMesssage{
			// 		Type: fmt.Sprintf("%v", parsed["type"]),
			// 		Data: parsed["data"],
			// 	})
			// }

			select {
			case WSChan <- parsed:
				// fmt.Printf("[WS] Queued signaling message: %v\n", parsed["message"])
			default:
				fmt.Printf("[WS] WSChan FULL - DROPPED: %v\n", parsed["message"])
			}
		}
	}()
}

// func HandleWebsocketMessage(msg map[string]any) tea.Msg {
// 	// msgType, _ := msg["type"].(string)
// 	// if msgType != "recieveMessage" {
// 	// 	return nil
// 	// }

// 	// dataMap, ok := msg["data"].(map[string]any)
// 	// if !ok {
// 	// 	return nil
// 	// }

// 	// name, _ := dataMap["name"].(string)
// 	// if name == "" {
// 	// 	name, _ = dataMap["name"].(string)
// 	// }

// 	// content, _ := dataMap["content"].(string)
// 	// serverID := fmt.Sprintf("%v", dataMap["serverID"])
// 	// channelID := fmt.Sprintf("%v", dataMap["channelID"])

// 	// var msgID int64
// 	// if idFloat, ok := dataMap["id"].(float64); ok {
// 	// 	msgID = int64(idFloat)
// 	// }

// 	// var timestamp int64
// 	// if tsFloat, ok := dataMap["timestamp"].(float64); ok {
// 	// 	timestamp = int64(tsFloat)
// 	// }

// 	// newMessage := Message{
// 	// 	ID:        msgID,
// 	// 	Username:  name,
// 	// 	Content:   content,
// 	// 	Timestamp: timestamp,
// 	// }

// 	// return WebsocketMsg{
// 	// 	ServerID:  serverID,
// 	// 	ChannelID: channelID,
// 	// 	Message:   newMessage,
// 	// }
// }

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

type WebsocketMesssage struct {
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

		// fmt.Println("i dont know what's happening")

		// if GlobalCallControl != nil {
		// 	GlobalCallControl.HandleSignalMessage(WebsocketMesssage{
		// 		Type: msgType,
		// 		Data: msg["data"],
		// 	})
		// }

		return WebsocketMesssage{
			Type: msgType,
			Data: msg["data"],
		}
	}
}
