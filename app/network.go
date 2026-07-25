package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GET(url string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

// directly encodes the destination result
func POST(payload any, authKey string, address string, response any) error {

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create request
	req, err := http.NewRequest(
		http.MethodPost,
		address,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authKey)

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("network response failed: %s", resp.Status)
	}

	// Decode JSON directly into the provided struct
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}

	return nil
}
