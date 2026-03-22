package controllers

import (
	"QUICK-READ-SYSTEM/models"
	"QUICK-READ-SYSTEM/services"
	"QUICK-READ-SYSTEM/utils"
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