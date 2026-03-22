package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
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
		UserID:       userID,
		VehicleType:  vehicleType,
		LicensePlate: licensePlate,
		IsAvailable:  true,
		Rating:       5.0,
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

// UpdateDriverLocation updates the driver's current GPS position
func (s *DeliveryService) UpdateDriverLocation(userID uint, lat, lng float64) error {
	return config.DB.Model(&models.DeliveryDriver{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{"current_lat": lat, "current_lng": lng}).Error
}

// AssignDriver finds the closest available driver and assigns them to an order
func (s *DeliveryService) AssignDriver(orderID string) (*models.DeliveryDriver, error) {
	// Get the order
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		return nil, errors.New("order not found")
	}

	if order.DriverID != 0 {
		return nil, errors.New("order already assigned to a driver")
	}

	// Get all available drivers
	var drivers []models.DeliveryDriver
	config.DB.Where("is_available = ?", true).Find(&drivers)

	if len(drivers) == 0 {
		return nil, errors.New("no available drivers at the moment")
	}

	// Find the closest driver (simple Haversine distance)
	var bestDriver *models.DeliveryDriver
	bestDist := math.MaxFloat64

	for i := range drivers {
		d := haversineDistance(
			drivers[i].CurrentLat, drivers[i].CurrentLng,
			order.DeliveryLat, order.DeliveryLng,
		)
		if d < bestDist {
			bestDist = d
			bestDriver = &drivers[i]
		}
	}

	// Assign driver to order
	order.DriverID = bestDriver.UserID
	order.Status = "pick_up"
	config.DB.Save(&order)

	// Mark driver as unavailable
	bestDriver.IsAvailable = false
	config.DB.Save(bestDriver)

	// Create route
	route := models.Route{
		OrderID:      order.ID,
		DriverID:     bestDriver.UserID,
		StartLat:     bestDriver.CurrentLat,
		StartLng:     bestDriver.CurrentLng,
		EndLat:       order.DeliveryLat,
		EndLng:       order.DeliveryLng,
		Distance:     bestDist,
		EstimatedMin: int(bestDist / 0.5), // rough estimate: 0.5 km/min
		Status:       "active",
	}
	config.DB.Create(&route)

	// Log delivery history
	s.LogDeliveryEvent(order.ID, bestDriver.UserID, "assigned", bestDriver.CurrentLat, bestDriver.CurrentLng, "Driver assigned to order")

	return bestDriver, nil
}

// LogDeliveryEvent records a delivery history entry
func (s *DeliveryService) LogDeliveryEvent(orderID, driverID uint, status string, lat, lng float64, note string) {
	history := models.DeliveryHistory{
		OrderID:   orderID,
		DriverID:  driverID,
		Status:    status,
		Latitude:  lat,
		Longitude: lng,
		Timestamp: time.Now(),
		Note:      note,
	}
	config.DB.Create(&history)
}

// GetDeliveryHistory returns all delivery history for a driver
func (s *DeliveryService) GetDeliveryHistory(driverID uint) ([]models.DeliveryHistory, error) {
	var history []models.DeliveryHistory
	if err := config.DB.Where("driver_id = ?", driverID).Order("created_at DESC").Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

func (s *DeliveryService) GetDriverStats(driverID uint) (map[string]int64, error) {
	var successful, pending, cancelled int64
	config.DB.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driverID, "delivered").Count(&successful)
	config.DB.Model(&models.Order{}).Where("driver_id = ? AND (status = ? OR status = ? OR status = ?)", driverID, "pick_up", "in_transit", "out_of_delivery").Count(&pending)
	config.DB.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driverID, "cancelled").Count(&cancelled)

	return map[string]int64{
		"successful": successful,
		"pending":    pending,
		"cancelled":  cancelled,
	}, nil
}

// GetOrderDeliveryHistory returns delivery history for a specific order
func (s *DeliveryService) GetOrderDeliveryHistory(orderID string) ([]models.DeliveryHistory, error) {
	var history []models.DeliveryHistory
	if err := config.DB.Where("order_id = ?", orderID).Order("created_at ASC").Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

// GetRoute returns the route for an order
func (s *DeliveryService) GetRoute(orderID string) (*models.Route, error) {
	var route models.Route
	if err := config.DB.Where("order_id = ?", orderID).First(&route).Error; err != nil {
		return nil, errors.New("route not found")
	}
	return &route, nil
}

// haversineDistance calculates distance between two lat/lng points in km
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
