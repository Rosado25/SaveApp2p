package models

import "time"

type Device struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	DeviceID   string    `json:"device_id"`
	UserID     int64     `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}
