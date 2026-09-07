package config

import (
	"database/sql"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
)

func ConnectDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err = db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	)`); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL UNIQUE,
		user_id INTEGER NOT NULL,
		name TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_devices_user_id ON devices(user_id)`); err != nil {
		_ = db.Close()
		return nil, err
	}

	fmt.Println("Connected to database")
	return db, nil
}
