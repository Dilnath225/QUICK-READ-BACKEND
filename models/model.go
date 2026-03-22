package models

import (
	"time"

	"gorm.io/gorm"
)

// Users (Patients, Pharmacists, Delivery)
type User struct {
	gorm.Model
	Name        string `json:"name"`
	Email       string `json:"email" gorm:"uniqueIndex;size:255"`
	Password    string `json:"-"`
	Role        string `json:"role" gorm:"index;size:100"` // "patient", "pharmacist", "delivery"
	Address        string `json:"address"`
	DateOfBirth    string `json:"date_of_birth"`
	Phone          string `json:"phone"`
	PushToken      string `json:"push_token"`
	ProfilePicture string `json:"profile_picture"`
	IsVerified     bool   `json:"is_verified" gorm:"default:false"`
	OTP            string `json:"-"` // Stored only when Redis is unavailable
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

type DeliveryDriver struct {
	gorm.Model
	UserID       uint    `json:"user_id" gorm:"uniqueIndex"`
	VehicleType  string  `json:"vehicle_type"` // "bike", "car", "van"
	LicensePlate string  `json:"license_plate"`
	IsAvailable  bool    `json:"is_available" gorm:"default:true"`
	CurrentLat   float64 `json:"current_lat"`
	CurrentLng   float64 `json:"current_lng"`
	Rating       float64 `json:"rating" gorm:"default:5.0"`
	TotalTrips   int     `json:"total_trips" gorm:"default:0"`
}

type DeliveryHistory struct {
	gorm.Model
	OrderID   uint      `json:"order_id" gorm:"index"`
	DriverID  uint      `json:"driver_id" gorm:"index"`
	Status    string    `json:"status"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
	Note      string    `json:"note"`
}

type Route struct {
	gorm.Model
	OrderID      uint    `json:"order_id" gorm:"index"`
	DriverID     uint    `json:"driver_id" gorm:"index"`
	StartLat     float64 `json:"start_lat"`
	StartLng     float64 `json:"start_lng"`
	EndLat       float64 `json:"end_lat"`
	EndLng       float64 `json:"end_lng"`
	Distance     float64 `json:"distance"`      // in kilometers
	EstimatedMin int     `json:"estimated_min"` // estimated minutes
	Status       string  `json:"status"`        // "active", "completed", "cancelled"
}

type Notification struct {
	gorm.Model
	UserID  uint   `json:"user_id" gorm:"index"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"` // "order_update", "inventory_alert", "delivery_update", "system"
	IsRead  bool   `json:"is_read" gorm:"default:false"`
}
