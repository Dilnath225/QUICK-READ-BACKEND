package controllers

import (
	"QUICK-READ-BACKEND/models"
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Input Structures
type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Address  string `json:"address" binding:"required"`
	Role     string `json:"role" binding:"required"` // "patient", "pharmacist", "rider"
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var authService = new(services.AuthService)

// POST /register
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := authService.Register(input.Name, input.Email, input.Password, input.Address, input.Role); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "email already exists" {
			status = http.StatusConflict
		}
		utils.ErrorResponse(c, status, err.Error())
		return
	}

	utils.SuccessResponse(c, "Registration successful.Please check your email for the OTP.", nil)
}

// POST /verify-otp
func VerifyOTP(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required"`
		OTP   string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := authService.VerifyOTP(input.Email, input.OTP); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, "OTP verified successfully. You can now login.", nil)
}

// POST /login
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	token, user, err := authService.Login(input.Email, input.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.SuccessResponse(c, "Login successful", gin.H{
		"token":  token,
		"role":   user.Role,
		"userId": user.ID,
	})
}

// GET /users/profile
func GetProfile(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated")
		return
	}
	currentUser := userVal.(models.User)

	utils.SuccessResponse(c, "Profile fetched", currentUser)
}

// PUT /users/profile — Patient updates their profile (UI: Profile screen SAVE button)
func UpdateProfile(c *gin.Context) {
	// Middleware sets "user" as models.User directly
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated")
		return
	}
	currentUser := userVal.(models.User)

	var input struct {
		Name      string `json:"name"`
		DOB       string `json:"date_of_birth"`
		Address   string `json:"address"`
		Phone     string `json:"phone"`
		Email     string `json:"email"`
		PushToken string `json:"push_token"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updatedUser, err := authService.UpdateProfile(
		currentUser.ID,
		input.Name, input.DOB, input.Address, input.Phone, input.Email, input.PushToken,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("✅ Profile updated for user ID: %d", currentUser.ID)
	utils.SuccessResponse(c, "Profile updated", updatedUser)
}

// POST /users/profile-picture — Upload a new profile picture via Base64 JSON
func UploadProfilePicture(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated")
		return
	}
	currentUser := userVal.(models.User)

	var input struct {
		Base64Image string `json:"base64_image" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request. Base64 encoded image is required.")
		return
	}

	// Remove data URI prefix if it exists (e.g. data:image/jpeg;base64,...)
	b64data := input.Base64Image
	parts := strings.SplitN(b64data, ",", 2)
	if len(parts) == 2 {
		b64data = parts[1]
	}

	imgData, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to decode base64 image")
		return
	}

	// Create directory if it doesn't exist
	uploadDir := "uploads/profiles"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	// Generate unique filename
	fileName := fmt.Sprintf("%d_%d.jpg", currentUser.ID, time.Now().Unix())
	filePath := filepath.Join(uploadDir, fileName)

	// Save the decoded binary as a file
	if err := os.WriteFile(filePath, imgData, 0644); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to write image to disk")
		return
	}

	// Create URL
	imageURL := fmt.Sprintf("/%s", filepath.ToSlash(filePath))

	updatedUser, err := authService.UpdateProfilePicture(currentUser.ID, imageURL)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("✅ Profile picture updated for user ID: %d", currentUser.ID)
	utils.SuccessResponse(c, "Profile picture updated successfully", updatedUser)
}

// DELETE /users/profile — User deletes their own account
func DeleteAccount(c *gin.Context) {
	// Middleware verifies the user
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Not authenticated")
		return
	}
	currentUser := userVal.(models.User)

	// Call the service layer to delete or deactivate the user
	if err := authService.DeleteUser(currentUser.ID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	log.Printf("🗑️ Account deleted for user ID: %d", currentUser.ID)
	utils.SuccessResponse(c, "Account successfully deleted", nil)
}
