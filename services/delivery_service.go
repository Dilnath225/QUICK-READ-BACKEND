package services

import (
	"QUICK-READ-SYSTEM/config"
	"QUICK-READ-SYSTEM/models"
	"errors"
	"math"
	"time"
)

type DeliveryService struct{}

// RegisterDriver creates a delivery driver profile for an existing user with role "delivery"
func (s *DeliveryService) RegisterDriver(userID uint, vehicleType, licensePlate string) (*models.DeliveryDriver, error) {
// Check if driver profile already exists
var existing models.DeliveryDriver
if err := config.DB.Where("user_id = ?", userID).First(&existing).Error; err == nil {
return nil, errors.New("driver profile already exists")
}