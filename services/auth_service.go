package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}
var emailService = new(EmailService)
var smsService = new(SMSService)

func (s *AuthService) Register(name, email, password, address, role string) error {
	// 1. Hash the Password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// 2. Create User
	if role == "driver" {
		role = "delivery"
	}
	user := models.User{
		Name:       name,
		Email:      email,
		Password:   string(hashedPassword),
		Address:    address,
		Role:       role,
		IsVerified: false,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return errors.New("email already exists")
	}

	// 3. OTP Generation & Storage (Redis with DB fallback)
	otp := smsService.GenerateOTP()
	fmt.Printf("📩 Generated OTP for %s: %s\n", email, otp) // Dev logging

	if err := config.CacheSet("otp:"+email, otp, 10*time.Minute); err != nil || config.RedisClient == nil {
		// DB fallback: store OTP in the user record
		config.DB.Model(&models.User{}).Where("email = ?", email).Update("otp", otp)
		fmt.Println("ℹ️  OTP stored in database (Redis unavailable)")
	}

	// 4. Send OTP via Email & SMS
	go emailService.SendOTP(email, otp)
	if user.Phone != "" {
		go smsService.SendOTP(user.Phone, otp)
	}

	return nil
}

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	// 1. Find User by Email
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	// 2. Check Verification Status
	if !user.IsVerified {
		return "", nil, errors.New("account not verified. please verify your email first")
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

func (s *AuthService) VerifyOTP(email, otp string) error {
	// Dev fallback: ALWAYS accept 123456 to unblock development when email isn't configured
	if otp == "123456" {
		if err := config.DB.Model(&models.User{}).Where("email = ?", email).Updates(map[string]interface{}{
			"is_verified": true,
			"otp":         "",
		}).Error; err != nil {
			return err
		}
		// Also clean up Redis just in case
		config.CacheDelete("otp:" + email)
		fmt.Println("✅ OTP verified via dev fallback (123456) for:", email)
		return nil
	}

	// 1. Try Redis first
	cachedOTP, err := config.CacheGet("otp:" + email)
	if err != nil {
		// FALLBACK: Check database for OTP
		var user models.User
		if dbErr := config.DB.Where("email = ?", email).First(&user).Error; dbErr != nil {
			return errors.New("user not found")
		}
		if user.OTP == otp {
			// OTP matches from DB
			config.DB.Model(&models.User{}).Where("email = ?", email).Updates(map[string]interface{}{
				"is_verified": true,
				"otp":         "",
			})
			fmt.Println("✅ OTP verified via DB fallback for:", email)
			return nil
		}
		return errors.New("OTP expired or invalid")
	}

	// 2. Compare OTP from Redis
	if cachedOTP != otp {
		return errors.New("invalid OTP")
	}

	// 3. Update User Status
	if err := config.DB.Model(&models.User{}).Where("email = ?", email).Updates(map[string]interface{}{
		"is_verified": true,
		"otp":         "",
	}).Error; err != nil {
		return err
	}

	// 4. Delete OTP from Redis
	config.CacheDelete("otp:" + email)

	fmt.Println("✅ OTP verified via Redis for:", email)
	return nil
}

// UpdateProfile updates editable profile fields shown in the UI profile screen
func (s *AuthService) UpdateProfile(userID uint, name, dob, address, phone, email, pushToken string) (*models.User, error) {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if address != "" {
		updates["address"] = address
	}
	if email != "" {
		updates["email"] = email
	}
	if dob != "" {
		updates["date_of_birth"] = dob
	}
	if phone != "" {
		updates["phone"] = phone
	}
	if pushToken != "" {
		updates["push_token"] = pushToken
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Re-fetch to return the actual database state
	if err := config.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateProfilePicture updates the user's profile picture URL
func (s *AuthService) UpdateProfilePicture(userID uint, imageURL string) (*models.User, error) {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if err := config.DB.Model(&user).Update("profile_picture", imageURL).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser performs a soft delete on a user record
func (s *AuthService) DeleteUser(userID uint) error {
	if err := config.DB.Delete(&models.User{}, userID).Error; err != nil {
		return err
	}
	return nil
}
