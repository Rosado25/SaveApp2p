package controllers

import (
	"Server/helpers"
	"Server/services"
	"encoding/json"
	"net/http"
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
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		helpers.JSON(w, 400, "Invalid request payload")
		return
	}
	if err := c.service.Login(w, credentials.Email, credentials.Password); err != nil {
		helpers.JSON(w, 401, "Invalid credentials")
		return
	}

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
