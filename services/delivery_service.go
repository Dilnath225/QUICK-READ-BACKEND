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

driver := models.DeliveryDriver{
UserID: userID,
VehicleType: vehicleType,
LicensePlate: licensePlate,
IsAvailable: true,
Rating: 5.0,
}

if err := config.DB.Create(&driver).Error; err != nil {
return nil, err
}

return &driver, nil
}

// GetDriverProfile returns the driver profile for a user
func (s *DeliveryService) GetDriverProfile(userID uint) (*models.DeliveryDriver, error) {
var driver models.DeliveryDriver
if err := config.DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
return nil, errors.New("driver profile not found")
}
return &driver, nil
}
// UpdateDriverProfile updates editable driver fields
func (s *DeliveryService) UpdateDriverProfile(userID uint, vehicleType, licensePlate string) (*models.DeliveryDriver, error) {
var driver models.DeliveryDriver
if err := config.DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
return nil, errors.New("driver profile not found")
}
updates := map[string]interface{}{}
if vehicleType != "" {
updates["vehicle_type"] = vehicleType
}
if licensePlate != "" {
updates["license_plate"] = licensePlate
}

if err := config.DB.Model(&driver).Updates(updates).Error; err != nil {
return nil, err
}

return &driver, nil
}
// ToggleAvailability sets the driver's availability status
func (s *DeliveryService) ToggleAvailability(userID uint, available bool) (*models.DeliveryDriver, error) {
var driver models.DeliveryDriver
if err := config.DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
return nil, errors.New("driver profile not found")
}

driver.IsAvailable = available
config.DB.Save(&driver)
return &driver, nil
}