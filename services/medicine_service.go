package services

import (
	"QUICK-READ-SYSTEM/config"
	"QUICK-READ-SYSTEM/models"
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
