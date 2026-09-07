package services

import (
	"Server/models"
	"Server/repositories"
	"errors"
)

type DeviceService struct {
	repo *repositories.UserRepository
}

func NewDeviceService(r *repositories.UserRepository) *DeviceService {
	return &DeviceService{repo: r}
}

func (s *DeviceService) CreateDevice(device *models.Device) error {
	if device.Name == "" {
		return errors.New("device name is required")
	}
	if device.DeviceID == "" {
		return errors.New("device id is required")
	}
	if device.UserID <= 0 {
		return errors.New("user id is required")
	}

	return s.repo.CreateDevice(device)
}

func (s *DeviceService) GetDevicesByUser(userID int64) ([]models.Device, error) {
	return s.repo.GetDevicesByUser(userID)
}
