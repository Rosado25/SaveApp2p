package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

var (
	tokenMu      sync.RWMutex
	currentToken string
)

func SaveTokenFromResponse(resp *http.Response) bool {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "token" && cookie.Value != "" {
			tokenMu.Lock()
			currentToken = cookie.Value
			tokenMu.Unlock()
			return true
		}
	}

	return false
}

func GetToken() string {
	tokenMu.RLock()
	defer tokenMu.RUnlock()
	return currentToken
}

func ClearToken() {
	tokenMu.Lock()
	currentToken = ""
	tokenMu.Unlock()
}

func ValidateToken() (string, error) {
	token := GetToken()
	if token == "" {
		return "", errors.New("missing token")
	}

	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/CheckToken", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("token is invalid or expired")
	}

	return token, nil
}
