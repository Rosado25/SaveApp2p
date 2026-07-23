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

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	)`); err != nil {
		_ = db.Close()
		return nil, err
	}

	fmt.Println("Connected to database")
	return db, nil
}
