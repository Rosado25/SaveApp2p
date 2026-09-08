package controllers

import (
	"Server/helpers"
	"Server/services"
	"encoding/json"
	"net/http"
	"strings"
)

type AuthController struct {
	service *services.AuthService
}

func NewAuthController(s *services.AuthService) *AuthController {
	return &AuthController{service: s}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {

	if helpers.HasValidSession(r) {
		helpers.JSON(w, http.StatusConflict, "Already logged in")
		return
	}

	if r.Method != http.MethodPost {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}
	client := helpers.ClientAddress(r)
	if !helpers.AllowLogin(client) {
		w.Header().Set("Retry-After", "900")
		helpers.JSON(w, http.StatusTooManyRequests, "Too many login attempts")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		DeviceID string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		helpers.JSON(w, 400, "Invalid request payload")
		return
	}
	credentials.Email = strings.TrimSpace(credentials.Email)
	credentials.DeviceID = strings.TrimSpace(credentials.DeviceID)
	if credentials.Email == "" || len(credentials.Email) > 320 || credentials.Password == "" || len(credentials.Password) > 1024 || len(credentials.DeviceID) > 128 {
		helpers.RecordLoginFailure(client)
		helpers.JSON(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	if err := c.service.Login(w, credentials.Email, credentials.Password, credentials.DeviceID); err != nil {
		helpers.RecordLoginFailure(client)
		helpers.JSON(w, 401, "Invalid credentials")
		return
	}
	helpers.ResetLoginFailures(client)

	helpers.JSON(w, 200, "Login successful")
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}

	if err := c.service.Logout(w); err != nil {
		helpers.JSON(w, 500, "Failed to logout")
		return
	}

	helpers.JSON(w, 200, "Logout successful")
}

func (c *AuthController) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}

	req, err := helpers.GetTokenFromCookie(r)
	if err != nil {
		helpers.JSON(w, 401, "Not logged in")
		return
	}
	helpers.JSON(w, 200, map[string]string{"token": req})
}

func (c *AuthController) CheckConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}

	helpers.JSON(w, 200, "Token is valid")
}
