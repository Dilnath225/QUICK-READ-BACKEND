package controllers

import (
	"QUICK-READ-SYSTEM/models"
	"QUICK-READ-SYSTEM/services"
	"QUICK-READ-SYSTEM/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

 AddToCart(c *gin.Context) {

	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var input struct {
		MedicineID uint `json:"medicine_id" binding:"required"`
		Quantity   int  `json:"quantity" binding:"required"`
	}