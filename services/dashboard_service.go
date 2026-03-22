package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"time"
)

type DashboardService struct{}

// RevenueReport holds daily revenue data
type RevenueReport struct {
	Date         string  `json:"date"`
	TotalRevenue float64 `json:"total_revenue"`
	OrderCount   int64   `json:"order_count"`
	DeliveryFees float64 `json:"delivery_fees"`
}
// OrderStats holds order statistics
type OrderStats struct {
	TotalOrders   int64 `json:"total_orders"`
	Processing    int64 `json:"processing"`
	ReadyToShip   int64 `json:"ready_to_ship"`
	InTransit     int64 `json:"in_transit"`
	Delivered     int64 `json:"delivered"`
	TodayOrders   int64 `json:"today_orders"`
	WeeklyOrders  int64 `json:"weekly_orders"`
	MonthlyOrders int64 `json:"monthly_orders"`
}

// GetDailyRevenue returns revenue report for a given pharmacy for the requested number of days
func (s *DashboardService) GetDailyRevenue(pharmacyID uint, days int) ([]RevenueReport, error) {
	var reports []RevenueReport

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		var totalRevenue float64
		var deliveryFees float64
		var count int64

		config.DB.Model(&models.Order{}).
			Where("pharmacy_id = ? AND created_at >= ? AND created_at < ? AND payment_status = ?",
				pharmacyID, startOfDay, endOfDay, "paid").
			Count(&count)

		config.DB.Model(&models.Order{}).
			Where("pharmacy_id = ? AND created_at >= ? AND created_at < ? AND payment_status = ?",
				pharmacyID, startOfDay, endOfDay, "paid").
			Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue)

		config.DB.Model(&models.Order{}).
			Where("pharmacy_id = ? AND created_at >= ? AND created_at < ? AND payment_status = ?",
				pharmacyID, startOfDay, endOfDay, "paid").
			Select("COALESCE(SUM(delivery_fee), 0)").Scan(&deliveryFees)

		reports = append(reports, RevenueReport{
			Date:         startOfDay.Format("2006-01-02"),
			TotalRevenue: totalRevenue,
			OrderCount:   count,
			DeliveryFees: deliveryFees,
		})
	}

	return reports, nil
}
// GetOrderStats returns order statistics for a pharmacy
func (s *DashboardService) GetOrderStats(pharmacyID uint) (*OrderStats, error) {
	stats := &OrderStats{}

	base := config.DB.Model(&models.Order{}).Where("pharmacy_id = ?", pharmacyID)

	base.Count(&stats.TotalOrders)
	config.DB.Model(&models.Order{}).Where("pharmacy_id = ? AND status = ?", pharmacyID, "processing").Count(&stats.Processing)
	config.DB.Model(&models.Order{}).Where("pharmacy_id = ? AND status = ?", pharmacyID, "ready_to_ship").Count(&stats.ReadyToShip)
	config.DB.Model(&models.Order{}).Where("pharmacy_id = ? AND status = ?", pharmacyID, "in_transit").Count(&stats.InTransit)
	config.DB.Model(&models.Order{}).Where("pharmacy_id = ? AND status = ?", pharmacyID, "delivered").Count(&stats.Delivered)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)
	monthStart := todayStart.AddDate(0, -1, 0)

	config.DB.Model(&models.Order{}).Where("pharmacy_id = ? AND created_at >= ?", pharmacyID, todayStart).Count(&stats.TodayOrders)
	config.DB.Model(&models.Order{}).Where("pharmacy_id = ? AND created_at >= ?", pharmacyID, weekStart).Count(&stats.WeeklyOrders)
	config.DB.Model(&models.Order{}).Where("pharmacy_id = ? AND created_at >= ?", pharmacyID, monthStart).Count(&stats.MonthlyOrders)

	return stats, nil
}
