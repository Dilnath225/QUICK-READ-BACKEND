package controllers

import (
	"QUICK-READ-BACKEND/models"
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

var orderService = new(services.OrderService)
var paymentService = new(services.PaymentService)

// POST /orders
func CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	createdOrder, err := orderService.CreateOrder(&order)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Order created", createdOrder)
}

func GetOrderHistory(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	orders, err := orderService.GetUserOrders(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Order history", orders)
}

// POST /orders/:id/checkout
func InitializeCheckout(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		PaymentMethod string `json:"payment_method" binding:"required"` // "card", "cod", "koko"
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	order, clientSecret, err := paymentService.InitializePayment(id, input.PaymentMethod)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Checkout initialized", gin.H{
		"order":         order,
		"client_secret": clientSecret, // For Stripe Frontend
	})
}
