package services

import (
	"Server/helpers"
	"Server/repositories"
	"errors"
	"net/http"
)

type AuthService struct {
	repo *repositories.UserRepository
}

func NewAuthService(r *repositories.UserRepository) *AuthService {
	return &AuthService{repo: r}
}

func (s *AuthService) Login(w http.ResponseWriter, email, password string) error {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}

	if !helpers.CheckPasswordHash(password, user.Password) {
		return errors.New("invalid credentials")
	}

	// Generate and set JWT token
	token, err := helpers.GenerateJWT(user.ID)
	if err != nil {
		return errors.New("failed to generate token")
	}
	helpers.SetTokenCookie(w, token)

	return nil
}

func (s *AuthService) Logout(w http.ResponseWriter) error {
	helpers.ClearTokenCookie(w)
	return nil
}
