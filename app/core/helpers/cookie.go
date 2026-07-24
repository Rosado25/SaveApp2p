package helpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

func SetToken(w http.ResponseWriter) {

	req, err := http.NewRequest("GET", "http://localhost:8080/Token", nil)
	if err != nil {
		JSON(w, 500, "Failed to create request")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    req.Header.Get("token"),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

}

func ValidateToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", err
	}

	body, err := json.Marshal(map[string]string{"token": cookie.Value})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/CheckToken", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	return cookie.Value, nil
}
