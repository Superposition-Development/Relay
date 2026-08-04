package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	mu sync.Mutex
)

type WSMessage struct {
	Message string      `json:"message"`
	AuthKey string      `json:"authKey"`
	Content interface{} `json:"content,omitempty"`
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

func RegisterWebsocket(address string, jwtToken string, messageHandler func(msg WSMessage)) {
	var err error

	socket, _, err = websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		fmt.Printf("WebSocket error: %v\n", err)
		return
	}

	registerMessage := WSMessage{
		Message: "register",
		AuthKey: jwtToken,
	}
	SendWebsocketJSON(registerMessage)

	go func() {
		defer func() {
			socket.Close()
			fmt.Println("Connection closed")
		}()

		for {
			_, messageData, err := socket.ReadMessage()
			if err != nil {
				fmt.Printf("WebSocket error or closed: %v\n", err)
				break
			}

			fmt.Printf("Message from server: %s\n", string(messageData))

			var parsed WSMessage
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

func SendWebsocketJSON(message WSMessage) {
	mu.Lock()
	defer mu.Unlock()

	if socket == nil {
		fmt.Println("Websocket offline")
		return
	}

	err := socket.WriteJSON(message)
	if err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
	}
}
