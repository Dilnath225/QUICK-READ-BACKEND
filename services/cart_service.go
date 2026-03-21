package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
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

// AddItem adds a medicine to the user's cart
func (s *CartService) AddItem(userID uint, medicineID uint, quantity int) (*models.Cart, error) {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// Get medicine details
	var medicine models.Medicine
	if err := config.DB.First(&medicine, medicineID).Error; err != nil {
		return nil, errors.New("medicine not found")
	}
	// Check stock
	if medicine.StockLevel < quantity {
		return nil, errors.New("insufficient stock")
	}
	// Check if item already exists in cart
	var existingItem models.CartItem
	err = config.DB.Where("cart_id = ? AND medicine_id = ?", cart.ID, medicineID).First(&existingItem).Error
	if err == nil {
		// Update quantity
		existingItem.Quantity += quantity
		config.DB.Save(&existingItem)
	} else {
		// Add new item
		item := models.CartItem{
			CartID:     cart.ID,
			MedicineID: medicineID,
			Name:       medicine.Name,
			Dosage:     medicine.Dosage,
			Price:      medicine.Price,
			Quantity:   quantity,
		}
		config.DB.Create(&item)
	}
	// Reload cart with items
	config.DB.Preload("Items").First(&cart, cart.ID)
	return cart, nil
}

// RemoveItem removes an item from the cart
func (s *CartService) RemoveItem(userID uint, itemID string) (*models.Cart, error) {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	result := config.DB.Where("id = ? AND cart_id = ?", itemID, cart.ID).Delete(&models.CartItem{})
	if result.RowsAffected == 0 {
		return nil, errors.New("item not found in cart")
	}

	config.DB.Preload("Items").First(&cart, cart.ID)
	return cart, nil
}

// ClearCart removes all items from the cart
func (s *CartService) ClearCart(userID uint) error {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return err
	}

	return config.DB.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error
}

// GetCartTotal calculates the total price of all items in the cart
func (s *CartService) GetCartTotal(cart *models.Cart) float64 {
	var total float64
	for _, item := range cart.Items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}
