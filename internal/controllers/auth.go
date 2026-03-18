package controllers

import (
	"QUICK-READ-BACKEND/internal/services"
	"QUICK-READ-BACKEND/internal/utils"
	"net/http"

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
