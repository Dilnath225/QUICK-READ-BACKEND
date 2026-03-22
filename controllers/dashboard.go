package controllers

import (
	"QUICK-READ-BACKEND/models"
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var dashboardService = new(services.DashboardService)

// GET /pharmacy/dashboard/revenue?days=7 — Daily revenue report
func GetDailyRevenue(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	days := 7 // Default to 7 days
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	reports, err := dashboardService.GetDailyRevenue(currentUser.ID, days)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Revenue report", gin.H{
		"pharmacy_id": currentUser.ID,
		"days":        days,
		"reports":     reports,
	})
}

// GET /pharmacy/dashboard/orders — Order statistics
func GetOrderStats(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	stats, err := dashboardService.GetOrderStats(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Order statistics", stats)
}
