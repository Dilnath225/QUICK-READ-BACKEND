package services

import (
	"QUICK-READ-BACKEND/models"
)

type AuthService struct{}

func (s *AuthService) Register(name, email, password, address, role string) error {
	return nil
}

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	return "dummy-token", &models.User{}, nil
}

func (s *AuthService) UpdateProfile(userID uint, nickName, dob, address, phone, email string) (*models.User, error) {
	return &models.User{}, nil
}
