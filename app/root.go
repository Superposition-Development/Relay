package app

import (
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"
)

type InteractionMode int

const (
	Default InteractionMode = iota
	Write
)

func (im InteractionMode) String() string {
	return stateName[im]
}

var stateName = map[InteractionMode]string{
	Default: "default",
	Write:   "write",
}

type Server struct {
	ID   any    `json:"id"`
	PFP  string `json:"pfp"`
	Name string `json:"name"`
}

type Channel struct {
	ID   any    `json:"id"`
	Name string `json:"name"`
}

type Message struct {
	ID        int64  `json:"id"`
	Username  string `json:"name"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	Pfp       string `json:"pfp"`
}

var currentServerAddress = ""
var CurrentInteractionMode = Default
var CurrentUserID = ""
var Socket *websocket.Conn
var ServerListToDataMap = make(map[int]Server, 0)
var ChannelListToDataMap = make(map[int]Channel, 0)
var Servers = make([]Server, 0)
var Channels = make([]Channel, 0)
var Messages = make([]Message, 0)

func SetCurrentServerAddress(address string) {
	currentServerAddress = address
}

func GetCurrentServerAddress() string {
	return currentServerAddress
}

func GetLocalFilesPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, "Relay")

	err = os.MkdirAll(appDir, 0700)
	if err != nil {
		return "", err
	}
	return appDir, nil
}

func ReverseMessages(msgs []Message) []Message {
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs
}

func AppendMessage(msg Message) {
	Messages = append(Messages, msg)
}

func PrependMessages(olderMsgs []Message) {
	reversed := ReverseMessages(olderMsgs)
	Messages = append(reversed, Messages...)
}
