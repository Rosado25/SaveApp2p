package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

var ErrDeviceSetupRequired = errors.New("device setup required")

func Login(email string, password string) (bool, error) {

	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		DeviceID string `json:"device_id,omitempty"`
	}
	credentials.Email = email
	credentials.Password = password
	deviceID, err := GetDeviceID()
	if err != nil {
		return false, err
	}
	credentials.DeviceID = deviceID

	body, err := json.Marshal(credentials)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/login", bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := apiClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, errors.New("invalid credentials")
	}

	if !SaveTokenFromResponse(resp) {
		return false, errors.New("server did not return a token")
	}

	// Keep the authenticated session open so the UI can ask for the device name.
	if deviceID == "" {
		return true, ErrDeviceSetupRequired
	}

	return true, nil
}

func Logout() bool {
	req, err := http.NewRequest("POST", "http://localhost:8080/logout", nil)
	if err != nil {
		return false
	}

	token := GetToken()
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
	}

	resp, err := apiClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	ClearToken()
	return true
}

func Register(email string, password string) bool {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	credentials.Email = email
	credentials.Password = password

	body, err := json.Marshal(credentials)
	if err != nil {
		return false
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/CreateUser", bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := apiClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusCreated
}
