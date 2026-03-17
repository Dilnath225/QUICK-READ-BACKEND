package controllers

// Input Structures
type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Address  string `json:"address" binding:"required"`
	Role     string `json:"role" binding:"required"` // "patient", "pharmacist", "rider"
}
