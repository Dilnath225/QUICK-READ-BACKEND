package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"errors"
	"fmt"
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

// DeleteMedicine removes a medicine from the database
func (s *MedicineService) DeleteMedicine(id string) error {
	result := config.DB.Delete(&models.Medicine{}, id)
	if result.RowsAffected == 0 {
		return errors.New("medicine not found")
	}

	// Invalidate medicine cache
	config.CacheDeletePattern("cache:*")

	return result.Error
}

// GetAll retrieves all medicines
func (s *MedicineService) GetAll() ([]models.Medicine, error) {
	var medicines []models.Medicine
	if err := config.DB.Find(&medicines).Error; err != nil {
		return nil, err
	}
	return medicines, nil
}

// Search retrieves medicines whose name matches the query
func (s *MedicineService) Search(query string) ([]models.Medicine, error) {
	var medicines []models.Medicine
	if err := config.DB.Where("name LIKE ?", "%"+query+"%").Find(&medicines).Error; err != nil {
		return nil, err
	}
	return medicines, nil
}

// GetPharmaciesForPrescription finds pharmacies that have the medicines in the prescription
func (s *MedicineService) GetPharmaciesForPrescription(medNames []string) ([]models.User, error) {
	var pharmacies []models.User
	if err := config.DB.Table("users").
		Select("users.*").
		Joins("JOIN medicines ON medicines.pharmacy_id = users.id").
		Where("users.role = ?", "pharmacist").
		Where("medicines.name IN ?", medNames).
		Group("users.id").
		Find(&pharmacies).Error; err != nil {
		return nil, err
	}
	return pharmacies, nil
}

// GetPriceList returns the prices of the medicines at a specific pharmacy
func (s *MedicineService) GetPriceList(pharmacyID uint, medNames []string) ([]models.Medicine, float64, error) {
	var medicines []models.Medicine
	if err := config.DB.Where("pharmacy_id = ? AND name IN ?", pharmacyID, medNames).Find(&medicines).Error; err != nil {
		return nil, 0, err
	}

	var total float64
	for _, m := range medicines {
		total += m.Price
	}

	return medicines, total, nil
}

// GetByID retrieves a single medicine by its ID
func (s *MedicineService) GetByID(id string) (*models.Medicine, error) {
	var medicine models.Medicine
	if err := config.DB.First(&medicine, id).Error; err != nil {
		return nil, fmt.Errorf("medicine not found")
	}
	return &medicine, nil
}
