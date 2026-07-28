package app

import (
	"os"
	"path/filepath"
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

var currentServerAddress = ""
var CurrentInteractionMode = Default

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
