package services

import (
	"QUICK-READ-BACKEND/models"
	"errors"
	"os"
	"time"

	"QUICK-READ-BACKEND/config"

	"github.com/golang-jwt/jwt/v5"
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

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	// 1. Find User by Email
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	// 2. Check Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	// 3. Generate JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
	})

	// Sign token with a secret key
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", nil, err
	}

	return tokenString, &user, nil
}
