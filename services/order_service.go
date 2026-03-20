package services

import (
	"QUICK-READ-SYSTEM/config"
	"QUICK-READ-SYSTEM/models"
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
