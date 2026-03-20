package services

import (
	"QUICK-READ-SYSTEM/config"
	"QUICK-READ-SYSTEM/models"
)

type NotificationService struct{}

// Create creates a new notification for a user
func (s *NotificationService) Create(userID uint, title, message, notifType string) (*models.Notification, error) {
	notif := models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
		IsRead:  false,
	}

	if err := config.DB.Create(&notif).Error; err != nil {
		return nil, err
	}

	return &notif, nil
}

// GetUserNotifications returns all notifications for a user
func (s *NotificationService) GetUserNotifications(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}
