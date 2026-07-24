package helpers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func Login(email string, password string) bool {

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

	req, err := http.NewRequest("POST", "http://localhost:8080/login", bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
