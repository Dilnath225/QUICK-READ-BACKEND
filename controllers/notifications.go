package controllers

import (
	"QUICK-READ-SYSTEM/models"
	"QUICK-READ-SYSTEM/services"
	"QUICK-READ-SYSTEM/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)
var notifService = new(services.NotificationService)

// GET /notifications — Get user's notifications
func GetNotifications(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	notifications, err := notifService.GetUserNotifications(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	unreadCount, _ := notifService.GetUnreadCount(currentUser.ID)

	utils.SuccessResponse(c, "Notifications retrieved", gin.H{
		"notifications": notifications,
		"unread_count":  unreadCount,
	})
}
