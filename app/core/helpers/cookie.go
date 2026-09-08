package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"
)

var (
	tokenMu      sync.RWMutex
	currentToken string
	apiClient    = newAPIClient()
)

func newAPIClient() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return &http.Client{Timeout: 10 * time.Second}
	}

	return &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}
}

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

	req, err := http.NewRequest("POST", apiURL("/CheckToken"), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := apiClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("token is invalid or expired")
	}

	return token, nil
}
