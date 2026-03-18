package controllers

import (
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
	"net/http"

	"QUICK-READ-BACKEND/models"

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
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, "Registration successful", nil)
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
		NickName string `json:"nick_name"`
		DOB      string `json:"date_of_birth"`
		Address  string `json:"address"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updatedUser, err := authService.UpdateProfile(
		currentUser.ID,
		input.NickName, input.DOB, input.Address, input.Phone, input.Email,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Profile updated", updatedUser)
}
