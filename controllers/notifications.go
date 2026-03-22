package controllers

import (
	"QUICK-READ-BACKEND/models"
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
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
// PUT /notifications/:id/read — Mark notification as read
func MarkNotificationRead(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)
	notifID := c.Param("id")

	if err := notifService.MarkAsRead(notifID, currentUser.ID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Notification not found")
		return
	}

	utils.SuccessResponse(c, "Notification marked as read", nil)
}
// PUT /notifications/read-all — Mark all notifications as read
func MarkAllNotificationsRead(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	if err := notifService.MarkAllAsRead(currentUser.ID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "All notifications marked as read", nil)
}
