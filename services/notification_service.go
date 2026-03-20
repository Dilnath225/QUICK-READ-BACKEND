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

// GetUnreadCount returns the count of unread notifications
func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	if err := config.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notifID string, userID uint) error {
	result := config.DB.Model(&models.Notification{}).Where("id = ? AND user_id = ?", notifID, userID).Update("is_read", true)
	if result.RowsAffected == 0 {
		return config.DB.Error
	}
	return result.Error
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	return config.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Update("is_read", true).Error
}

// --- Convenience methods to send specific notification types ---

// NotifyOrderUpdate sends an order update notification
func (s *NotificationService) NotifyOrderUpdate(userID uint, orderID uint, status string) {
	s.Create(userID, "Order Update", "Your order #"+uintToStr(orderID)+" status: "+status, "order_update")
}

// NotifyInventoryAlert sends a low-stock alert to a pharmacist
func (s *NotificationService) NotifyInventoryAlert(pharmacistID uint, medicineName string, stockLevel int) {
	s.Create(pharmacistID, "Low Stock Alert", medicineName+" stock is low ("+intToStr(stockLevel)+" remaining)", "inventory_alert")
}

// NotifyDeliveryUpdate sends a delivery update to the customer
func (s *NotificationService) NotifyDeliveryUpdate(userID uint, message string) {
	s.Create(userID, "Delivery Update", message, "delivery_update")
}

func uintToStr(n uint) string {
	s := ""
	if n == 0 {
		return "0"
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func intToStr(n int) string {
	if n < 0 {
		return "-" + uintToStr(uint(-n))
	}
	return uintToStr(uint(n))
}
