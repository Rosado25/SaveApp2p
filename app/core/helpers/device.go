package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const deviceFileName = "device_id"

type deviceResponse struct {
	DeviceID string `json:"device_id"`
}

func deviceIDPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "SaveApp", deviceFileName), nil
}

func GetDeviceID() (string, error) {
	path, err := deviceIDPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func SaveDeviceID(deviceID string) error {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return errors.New("device ID is empty")
	}

	path, err := deviceIDPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(deviceID+"\n"), 0600)
}

func CreateDevice(name string) (string, error) {
	requestBody, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL("/CreateDevice"), bytes.NewReader(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := GetToken(); token != "" {
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
	}

	resp, err := apiClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create device failed: %s", resp.Status)
	}

	var device deviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return "", err
	}
	if err := SaveDeviceID(device.DeviceID); err != nil {
		return "", err
	}

	return device.DeviceID, nil
}

func SendTestMessage(targetDeviceID string, message string) error {
	return SendWebSocketMessage(
		targetDeviceID,
		"test",
		message,
	)
}
