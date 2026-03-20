package services

import (
	"QUICK-READ-SYSTEM/config"
	"QUICK-READ-SYSTEM/models"
	"errors"
)

type MedicineService struct{}

// AddMedicine creates a new medicine entry for a pharmacy
func (s *MedicineService) AddMedicine(name, dosage string, price float64, stockLevel int, pharmacyID uint) (*models.Medicine, error) {
	medicine := models.Medicine{
		Name:       name,
		Dosage:     dosage,
		Price:      price,
		StockLevel: stockLevel,
		PharmacyID: pharmacyID,
	}

	if err := config.DB.Create(&medicine).Error; err != nil {
		return nil, err
	}

	// Check if stock is low, fire event
	if stockLevel <= 10 && Bus != nil {
		Bus.Publish(Event{
			Type: EventStockLow,
			Payload: map[string]interface{}{
				"pharmacist_id": pharmacyID,
				"medicine_name": name,
				"stock_level":   stockLevel,
			},
		})
	}

	return &medicine, nil
}

// UpdateMedicine updates an existing medicine's details
func (s *MedicineService) UpdateMedicine(id string, name, dosage string, price float64, stockLevel int) (*models.Medicine, error) {
	var medicine models.Medicine
	if err := config.DB.First(&medicine, id).Error; err != nil {
		return nil, errors.New("medicine not found")
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if dosage != "" {
		updates["dosage"] = dosage
	}
	if price > 0 {
		updates["price"] = price
	}
	if stockLevel >= 0 {
		updates["stock_level"] = stockLevel
	}

	if err := config.DB.Model(&medicine).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Check if updated stock is low
	if stockLevel > 0 && stockLevel <= 10 && Bus != nil {
		Bus.Publish(Event{
			Type: EventStockLow,
			Payload: map[string]interface{}{
				"pharmacist_id": medicine.PharmacyID,
				"medicine_name": medicine.Name,
				"stock_level":   stockLevel,
			},
		})
	}

	// Invalidate medicine cache
	config.CacheDeletePattern("cache:*")

	return &medicine, nil
}
