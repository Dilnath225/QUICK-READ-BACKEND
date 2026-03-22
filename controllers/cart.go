package controllers

import (
	"QUICK-READ-BACKEND/models"
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

var cartService = new(services.CartService)

// POST /cart — Add item to cart
func AddToCart(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var input struct {
		MedicineID uint `json:"medicine_id" binding:"required"`
		Quantity   int  `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	cart, err := cartService.AddItem(currentUser.ID, input.MedicineID, input.Quantity)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Item added to cart", cart)
}

// GET /cart — Retrieve user's cart
func GetCart(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	cart, err := cartService.GetOrCreateCart(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Cart retrieved", cart)
}

// DELETE /cart/:id — Remove item from cart
func RemoveFromCart(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	itemID := c.Param("id")

	cart, err := cartService.RemoveItem(currentUser.ID, itemID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Item removed from cart", cart)
}

// DELETE /cart — Clear entire cart
func ClearCart(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	if err := cartService.ClearCart(currentUser.ID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Cart cleared", nil)
}
