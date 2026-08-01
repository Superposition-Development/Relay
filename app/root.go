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

var currentServerAddress = ""
var CurrentInteractionMode = Default
var CurrentUserID = ""
var socket *websocket.Conn
var ServerListToDataMap = make(map[int]Server, 0)
var ChannelListToDataMap = make(map[int]Channel, 0)

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
