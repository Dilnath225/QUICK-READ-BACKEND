package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"errors"
)

type OrderService struct{}

func (s *OrderService) CreateOrder(order *models.Order) (*models.Order, error) {
	// Report Feature: Emergency Order Handling
	if order.IsEmergency {
		// Logic to prioritize or assign to nearest pharmacy instantly
		order.Status = "priority_processing"
	} else {
		order.Status = "processing"
	}

	if err := config.DB.Create(order).Error; err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetOrder(id string) (*models.Order, error) {
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		return nil, errors.New("order not found")
	}
	return &order, nil
}

// GetUserOrders retrieves the order history for a patient
func (s *OrderService) GetUserOrders(userID uint) ([]models.Order, error) {
	var orders []models.Order
	if err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// GetDriverOrders retrieves the active orders for a delivery driver
func (s *OrderService) GetDriverOrders(driverID uint) ([]models.Order, error) {
	var orders []models.Order
	if err := config.DB.Where("driver_id = ? AND status != ?", driverID, "delivered").Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// GetPharmacyOrders retrieves all orders for a specific pharmacy
func (s *OrderService) GetPharmacyOrders(pharmacyID uint) ([]models.Order, error) {
	var orders []models.Order
	if err := config.DB.Where("pharmacy_id = ?", pharmacyID).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) UpdateLocation(id string, lat, lng float64) error {
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		return errors.New("order not found")
	}

	order.DeliveryLat = lat
	order.DeliveryLng = lng

	return config.DB.Save(&order).Error
}

// UpdateStatus moves an order through the 5 UI delivery stages:
// processing → ready_to_ship → pick_up → in_transit → out_of_delivery → review → delivered
func (s *OrderService) UpdateStatus(id string, newStatus string) (*models.Order, error) {
	validStatuses := map[string]bool{
		"processing":      true,
		"ready_to_ship":   true,
		"pick_up":         true,
		"in_transit":      true,
		"out_of_delivery": true,
		"review":          true,
		"delivered":       true,
	}

	if !validStatuses[newStatus] {
		return nil, errors.New("invalid status: must be one of processing, ready_to_ship, pick_up, in_transit, out_of_delivery, review, delivered")
	}

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		return nil, errors.New("order not found")
	}

	order.Status = newStatus
	if err := config.DB.Save(&order).Error; err != nil {
		return nil, err
	}

	// Trigger Notification
	notifService := &NotificationService{}
	notifService.NotifyOrderUpdate(order.UserID, order.ID, newStatus)

	// Publish delivery completion event
	if newStatus == "delivered" && Bus != nil {
		Bus.Publish(Event{
			Type: EventDeliveryCompleted,
			Payload: map[string]interface{}{
				"order_id": order.ID,
				"user_id":  order.UserID,
			},
		})
	}

	return &order, nil
}
