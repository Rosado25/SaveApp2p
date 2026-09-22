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
	if r.Method != http.MethodPost {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var request struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		helpers.JSON(w, 400, "Invalid request payload")
		return
	}

	DeviceID, err := helpers.GenerateDeviceID()
	if err != nil {
		helpers.JSON(w, 500, "Failed to generate device ID")
		return
	}

	userID, ok := helpers.AuthenticatedUserID(r.Context())
	if !ok {
		helpers.JSON(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	device := models.Device{
		Name:     request.Name,
		DeviceID: DeviceID,
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

func (c *DeviceController) GetDevicesByUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.JSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := helpers.AuthenticatedUserID(r.Context())
	if !ok {
		helpers.JSON(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	devices, err := c.service.GetDevicesByUser(userID)
	if err != nil {
		helpers.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.JSON(w, http.StatusOK, devices)
}

func (c *DeviceController) GetDeviceByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.JSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	deviceID := r.URL.Query().Get("id")
	if deviceID == "" {
		helpers.JSON(w, http.StatusBadRequest, "device_id is required")
		return
	}

	device, err := c.service.GetDeviceByID(deviceID)
	if err != nil {
		helpers.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.JSON(w, http.StatusOK, device)
}
