package services

import (
	"QUICK-READ-BACKEND/models"
	"errors"

	"QUICK-READ-BACKEND/config"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func (s *AuthService) Register(name, email, password, address, role string) error {
	//  Hash the Password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	//  Create User
	if role == "driver" {
		role = "delivery"
	}
	user := models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Address:  address,
		Role:     role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return errors.New("email already exists")
	}

	return nil
}
