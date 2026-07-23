package services

import (
	"Server/helpers"
	"Server/models"
	"Server/repositories"
	"errors"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(r *repositories.UserRepository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) CreateUser(user models.User) error {
	if user.Email == "" {
		return errors.New("email is required")
	}
	if user.Password == "" {
		return errors.New("password is required")
	}

	passwordHash := helpers.HashPassword(user.Password)
	user.Password = passwordHash

	return s.repo.CreateUser(user)
}

func (s *UserService) GetUsers() ([]models.User, error) {
	return s.repo.GetUsers()
}

func (s *UserService) GetUserByID(id string) (models.User, error) {
	if id == "" {
		return models.User{}, errors.New("id is required")
	}
	return s.repo.GetUserByID(id)
}

func (s *UserService) GetUserByEmail(email string) (models.User, error) {
	if email == "" {
		return models.User{}, errors.New("email is required")
	}
	return s.repo.GetUserByEmail(email)
}
