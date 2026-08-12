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

func RegisterWebsocket(address string, jwtToken string, messageHandler func(msg map[string]any)) {
	formattedAddr := address
	if !strings.HasPrefix(formattedAddr, "ws://") && !strings.HasPrefix(formattedAddr, "wss://") {
		formattedAddr = "ws://" + formattedAddr
	}

	conn, resp, err := websocket.DefaultDialer.Dial(formattedAddr, nil)
	if err != nil {
		if resp != nil {
			fmt.Printf("WebSocket dial failed with status %s: %v\n", resp.Status, err)
		} else {
			fmt.Printf("WebSocket dial error: %v\n", err)
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
	GlobalCallControl = InstantiateCallControl()

	go func() {
		defer func() {
			mu.Lock()
			if Socket != nil {
				Socket.Close()
				Socket = nil
			}
			mu.Unlock()
			fmt.Println("WebSocket connection closed")
		}()

		for {
			c := GetConn()
			if c == nil {
				break
			}

			_, messageData, err := c.ReadMessage()
			if err != nil {
				fmt.Printf("WebSocket read error: %v\n", err)
				break
			}

			var parsed map[string]any
			if err := json.Unmarshal(messageData, &parsed); err != nil {
				fmt.Printf("Failed to parse incoming JSON: %v\n", err)
				continue
			}

			if messageHandler != nil {
				messageHandler(parsed)
			}
		}
	}()
}

func SendWebsocketJSON(message any) {
	mu.Lock()
	defer mu.Unlock()

	if Socket == nil {
		fmt.Println("Websocket offline")
		return
	}

	err := Socket.WriteJSON(message)
	if err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
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

		msgType, ok := msg["type"].(string)
		if !ok {
			return nil
		}

		GlobalCallControl.HandleSignalMessage(WebsocketMesssage{
			Type: msgType,
			Data: msg["data"],
		})

		return WebsocketMesssage{
			Type: msgType,
			Data: msg["data"],
		}
	}
}
