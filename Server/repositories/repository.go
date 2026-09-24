package repositories

import (
	"Server/models"
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrDeviceAlreadyExists = errors.New("device already exists")
	ErrUserNotFound        = errors.New("user not found")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user models.User) error {
	query := `INSERT INTO users (name, email, password) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, user.Name, user.Email, user.Password)
	return err
}

func (r *UserRepository) GetUsers() ([]models.User, error) {
	query := `SELECT id, name, email, password FROM users ORDER BY id`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetUserByID(id string) (models.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var user models.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (models.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE email = ?`
	row := r.db.QueryRow(query, email)

	var user models.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) CreateDevice(device *models.Device) error {
	var userExists int
	if err := r.db.QueryRow(`SELECT 1 FROM users WHERE id = ?`, device.UserID).Scan(&userExists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	var deviceExists int
	if err := r.db.QueryRow(`SELECT 1 FROM devices WHERE device_id = ?`, device.DeviceID).Scan(&deviceExists); err == nil {
		return ErrDeviceAlreadyExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	query := `INSERT INTO devices (name, device_id, user_id) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, device.Name, device.DeviceID, device.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: devices.device_id") {
			return ErrDeviceAlreadyExists
		}
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		device.ID = int(id)
	}
	return nil
}

func (r *UserRepository) GetDevicesByUser(userID int64) ([]models.Device, error) {
	query := `SELECT id, name, device_id, user_id, created_at, last_seen_at
		FROM devices WHERE user_id = ? ORDER BY id`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var device models.Device
		if err := rows.Scan(
			&device.ID,
			&device.Name,
			&device.DeviceID,
			&device.UserID,
			&device.CreatedAt,
			&device.LastSeenAt,
		); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *UserRepository) DeviceBelongsToUser(userID int64, deviceID string) (bool, error) {
	var exists int
	err := r.db.QueryRow(
		`SELECT 1 FROM devices WHERE user_id = ? AND device_id = ?`,
		userID,
		deviceID,
	).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	_, err = r.db.Exec(
		`UPDATE devices SET last_seen_at = CURRENT_TIMESTAMP WHERE user_id = ? AND device_id = ?`,
		userID,
		deviceID,
	)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *UserRepository) GetDeviceByID(deviceID string) (*models.Device, error) {
	query := `SELECT id, name, device_id, user_id, created_at, last_seen_at FROM devices WHERE id = ?`
	row := r.db.QueryRow(query, deviceID)

	var device models.Device
	if err := row.Scan(
		&device.ID,
		&device.Name,
		&device.DeviceID,
		&device.UserID,
		&device.CreatedAt,
		&device.LastSeenAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("device not found")
		}
		return nil, err
	}

	return &device, nil
}

func (r *UserRepository) GetDevicesFromUser(userID int64) ([]models.Device, error) {
	query := `SELECT id, name, last_seen_at FROM devices WHERE user_id = ? ORDER BY id`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var device models.Device
		if err := rows.Scan(
			&device.ID,
			&device.Name,
			&device.LastSeenAt,
		); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return devices, nil
}

func (r *UserRepository) DeviceIdToName(deviceID string) (*models.Device, error) {
	query := `SELECT name FROM devices WHERE device_id = ?`
	row := r.db.QueryRow(query, deviceID)

	var device models.Device
	if err := row.Scan(&device.Name); err != nil {
		if err == sql.ErrNoRows {
			return &models.Device{}, errors.New("user not found")
		}
		return &models.Device{}, err
	}
	return &device, nil
}
