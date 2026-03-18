package models

import (
	"gorm.io/gorm"
)

// User Model (Patients, Pharmacists, Riders)
type User struct {
	gorm.Model
	Name     string `json:"name"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"`
	Address  string `json:"address"`
	Role     string `json:"role"` // "patient", "pharmacist", "rider"
}

// Medicines (for Stock Checking)
type Medicine struct {
	gorm.Model
	Name       string  `json:"name" gorm:"index;size:255"`
	Dosage     string  `json:"dosage"`
	Price      float64 `json:"price"`
	StockLevel int     `json:"stock_level"`
	PharmacyID uint    `json:"pharmacy_id" gorm:"index"`
}

// Prescriptions (Links to OCR & Verification)
type Prescription struct {
	gorm.Model
	UserID       uint   `json:"user_id" gorm:"index"`
	ImageURL     string `json:"image_url"`
	OCRText      string `json:"ocr_text"`
	VerifiedText string `json:"verified_text"`
	Status       string `json:"status" gorm:"index;size:50"` // "pending", "verified", "rejected"
}

// Orders (Includes Emergency Logic)
type Order struct {
	gorm.Model
	PrescriptionID uint    `json:"prescription_id"`
	UserID         uint    `json:"user_id" gorm:"index"`
	TotalAmount    float64 `json:"total_amount"`
	IsEmergency    bool    `json:"is_emergency"`
	Status         string  `json:"status" gorm:"index;size:50"`
	DeliveryLat    float64 `json:"delivery_lat"`
	DeliveryLng    float64 `json:"delivery_lng"`

	// Payment Fields
	PaymentMethod string  `json:"payment_method"` // "card", "cash", or "koko"
	PaymentStatus string  `json:"payment_status"` // "pending", "paid", "failed"
	TransactionID string  `json:"transaction_id"` // Stripe PaymentIntent ID
	DeliveryFee   float64 `json:"delivery_fee"`
	PharmacyID    uint    `json:"pharmacy_id" gorm:"index"`

	// Delivery Assignment
	DriverID uint `json:"driver_id" gorm:"index"`
}

type LocationUpdate struct {
	Lat float64 `json:"lat" binding:"required"`
	Lng float64 `json:"lng" binding:"required"`
}

// cart and cart items

type Cart struct {
	gorm.Model
	UserID uint       `json:"user_id" gorm:"uniqueIndex"`
	Items  []CartItem `json:"items" gorm:"foreignKey:CartID"`
}

type CartItem struct {
	gorm.Model
	CartID     uint    `json:"cart_id" gorm:"index"`
	MedicineID uint    `json:"medicine_id"`
	Name       string  `json:"name"`
	Dosage     string  `json:"dosage"`
	Price      float64 `json:"price"`
	Quantity   int     `json:"quantity"`
}
