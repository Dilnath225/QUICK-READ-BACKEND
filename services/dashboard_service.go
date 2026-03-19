package services

import (
	"QUICK-READ-SYSTEM/config"
	"QUICK-READ-SYSTEM/models"
	"time"
)

type DashboardService struct{}

// RevenueReport holds daily revenue data
type RevenueReport struct {
	Date         string  json:"date"
	TotalRevenue float64 json:"total_revenue"
	OrderCount   int64   json:"order_count"
	DeliveryFees float64 json:"delivery_fees"
}
