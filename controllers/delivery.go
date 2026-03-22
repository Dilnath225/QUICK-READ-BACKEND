package controllers

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
	"net/http"

	"log"

	"github.com/gin-gonic/gin"
)

var deliveryService = new(services.DeliveryService)
var orderServiceForDelivery = new(services.OrderService)

// GET /delivery/orders — Driver's active orders
func GetActiveOrders(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	log.Printf("DEBUG: GetActiveOrders called for user ID: %d, role: %s", currentUser.ID, currentUser.Role)

	orders, err := orderServiceForDelivery.GetDriverOrders(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Active orders", orders)
}

// PUT /orders/:id/location — Driver sends their new GPS location
func UpdateLocation(c *gin.Context) {
	orderID := c.Param("id")

	var input struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	order.DeliveryLat = input.Lat
	order.DeliveryLng = input.Lng
	order.Status = "in_transit"
	config.DB.Save(&order)

	// Broadcast to WebSocket clients
	BroadcastLocation(orderID, input.Lat, input.Lng, order.Status)

	// Log delivery event
	deliveryService.LogDeliveryEvent(order.ID, order.DriverID, "location_update", input.Lat, input.Lng, "Driver location updated")

	c.JSON(http.StatusOK, gin.H{"message": "Location updated", "status": order.Status})
}

// POST /delivery/register — Driver onboarding
func RegisterDriver(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var input struct {
		VehicleType  string `json:"vehicle_type" binding:"required"`
		LicensePlate string `json:"license_plate" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	driver, err := deliveryService.RegisterDriver(currentUser.ID, input.VehicleType, input.LicensePlate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, "Driver registered successfully", driver)
}

// GET /delivery/profile — Get driver profile
func GetDriverProfile(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	driver, err := deliveryService.GetDriverProfile(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, "Driver profile", driver)
}

// PUT /delivery/profile — Update driver profile
func UpdateDriverProfile(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var input struct {
		VehicleType  string `json:"vehicle_type"`
		LicensePlate string `json:"license_plate"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	driver, err := deliveryService.UpdateDriverProfile(currentUser.ID, input.VehicleType, input.LicensePlate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, "Driver profile updated", driver)
}

// POST /delivery/availability — Toggle driver availability
func ToggleAvailability(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var input struct {
		Available bool `json:"available"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	driver, err := deliveryService.ToggleAvailability(currentUser.ID, input.Available)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, "Availability updated", driver)
}

// POST /delivery/location — Update driver's current GPS position
func UpdateDriverLocation(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var input struct {
		Lat float64 `json:"lat" binding:"required"`
		Lng float64 `json:"lng" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := deliveryService.UpdateDriverLocation(currentUser.ID, input.Lat, input.Lng); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Driver location updated", nil)
}

// POST /delivery/assign/:orderId — Auto-assign closest available driver
func AssignDriver(c *gin.Context) {
	orderID := c.Param("orderId")

	driver, err := deliveryService.AssignDriver(orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Publish event
	if services.Bus != nil {
		var order models.Order
		config.DB.First(&order, orderID)
		services.Bus.Publish(services.Event{
			Type: services.EventDeliveryAssigned,
			Payload: map[string]interface{}{
				"user_id":   order.UserID,
				"order_id":  order.ID,
				"driver_id": driver.UserID,
			},
		})
	}

	utils.SuccessResponse(c, "Driver assigned", driver)
}

// GET /delivery/history — Driver's delivery history
func GetDeliveryHistory(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	history, err := deliveryService.GetDeliveryHistory(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Delivery history", history)
}

// GET /delivery/route/:orderId — Get route for an order
func GetRoute(c *gin.Context) {
	orderID := c.Param("orderId")

	route, err := deliveryService.GetRoute(orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, "Route details", route)
}

// GET /delivery/order-history/:orderId — Get delivery events for an order
func GetOrderDeliveryHistory(c *gin.Context) {
	orderID := c.Param("orderId")

	history, err := deliveryService.GetOrderDeliveryHistory(orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Order delivery history", history)
}

// GET /delivery/stats
func GetDriverStats(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	stats, err := deliveryService.GetDriverStats(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Driver statistics", stats)
}
