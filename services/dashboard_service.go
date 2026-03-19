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
// OrderStats holds order statistics
type OrderStats struct {
	TotalOrders   int64 json:"total_orders"
	Processing    int64 json:"processing"
	ReadyToShip   int64 json:"ready_to_ship"
	InTransit     int64 json:"in_transit"
	Delivered     int64 json:"delivered"
	TodayOrders   int64 json:"today_orders"
	WeeklyOrders  int64 json:"weekly_orders"
	MonthlyOrders int64 json:"monthly_orders"
}
