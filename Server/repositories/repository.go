package repositories

import (
	"Server/models"
	"database/sql"
	"errors"
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
