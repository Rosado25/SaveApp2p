package services

import (
	"Server/models"
	"Server/repositories"
	"errors"
	"strings"
)

type DeviceService struct {
	repo *repositories.UserRepository
}

func NewDeviceService(r *repositories.UserRepository) *DeviceService {
	return &DeviceService{repo: r}
}

func (s *DeviceService) CreateDevice(device *models.Device) error {
	device.Name = strings.TrimSpace(device.Name)
	if device.Name == "" {
		return errors.New("device name is required")
	}
	if len(device.Name) > 100 {
		return errors.New("device name is too long")
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

// DeviceBelongsToUser verifies the device identity presented when opening a
// persistent connection. It also records that the device was recently seen.
func (s *DeviceService) DeviceBelongsToUser(userID int64, deviceID string) (bool, error) {
	return s.repo.DeviceBelongsToUser(userID, deviceID)
}

func (s *DeviceService) GetDeviceByID(deviceID string) (*models.Device, error) {
	return s.repo.GetDeviceByID(deviceID)
}

func (s *DeviceService) GetDevicesFromUser(userID int64) ([]models.Device, error) {
	return s.repo.GetDevicesFromUser(userID)
}

func (s *DeviceService) DeviceIdToName(deviceID string) (*models.Device, error) {
	return s.repo.DeviceIdToName(deviceID)
}
