package services

import (
	"QUICK-READ-SYSTEM/config"
	"QUICK-READ-SYSTEM/models"
	"errors"
)
type CartService struct{}

// GetOrCreateCart finds the user's cart or creates a new one
func (s *CartService) GetOrCreateCart(userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := config.DB.Where("user_id = ?", userID).Preload("Items").First(&cart).Error
	if err != nil {
		// Create new cart
		cart = models.Cart{UserID: userID}
		if err := config.DB.Create(&cart).Error; err != nil {
			return nil, err
		}
	}
	return &cart, nil
}
