package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AuthData struct {
	Token string `json:"token"`
}

func GetAuthPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, "Relay")

	err = os.MkdirAll(appDir, 0700)
	if err != nil {
		return "", err
	}

	return filepath.Join(appDir, "auth.json"), nil
}

func SaveToken(token string) error {
	path, err := GetAuthPath()
	if err != nil {
		return err
	}

	data := AuthData{
		Token: token,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(
		path,
		jsonData,
		0600,
	)
}

func LoadToken() (string, error) {
	path, err := GetAuthPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var auth AuthData

	err = json.Unmarshal(data, &auth)
	if err != nil {
		return "", err
	}

	return auth.Token, nil
}

func ClearToken() error {
	path, err := GetAuthPath()
	if err != nil {
		return err
	}

	return os.Remove(path)
}
