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
