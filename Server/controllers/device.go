package controllers

import (
	"Server/helpers"
	"Server/models"
	"Server/repositories"
	"Server/services"
	"encoding/json"
	"errors"
	"net/http"
)

type DeviceController struct {
	service *services.DeviceService
}

func NewDeviceController(s *services.DeviceService) *DeviceController {
	return &DeviceController{service: s}
}

func (c *DeviceController) CreateDevice(w http.ResponseWriter, r *http.Request) {
	if helpers.CheckConnection(w, r, 1, "Not logged in") {
		return
	}

	if r.Method != http.MethodPost {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}

	var request struct {
		Name     string `json:"name"`
		DeviceID string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		helpers.JSON(w, 400, "Invalid request payload")
		return
	}

	token, err := helpers.GetTokenFromCookie(r)
	if err != nil {
		helpers.JSON(w, 401, "Not logged in")
		return
	}
	userID, err := helpers.ValidateJWT(token)
	if err != nil {
		helpers.JSON(w, 401, "Not logged in")
		return
	}

	device := models.Device{
		Name:     request.Name,
		DeviceID: request.DeviceID,
		UserID:   userID,
	}
	if err := c.service.CreateDevice(&device); err != nil {
		switch {
		case errors.Is(err, repositories.ErrDeviceAlreadyExists):
			helpers.JSON(w, http.StatusConflict, err.Error())
		case errors.Is(err, repositories.ErrUserNotFound):
			helpers.JSON(w, http.StatusNotFound, err.Error())
		default:
			helpers.JSON(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	helpers.JSON(w, 201, device)
}
