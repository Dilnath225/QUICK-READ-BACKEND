package services

import (
	"strconv"
	"testing"

	"QUICK-READ-BACKEND/config"
	"github.com/stretchr/testify/assert"
)

func TestMedicineService_CRUD(t *testing.T) {
	// Clean the medicine table for isolation
	config.DB.Exec("DELETE FROM medicines")

	medicineService := new(MedicineService)

	// --- 1. Test AddMedicine ---
	med, err := medicineService.AddMedicine("Amoxicillin", "250mg", 120.50, 50, 1)
	assert.NoError(t, err, "Adding medicine should succeed")
	assert.NotNil(t, med)
	assert.Equal(t, "Amoxicillin", med.Name)
	assert.Equal(t, "250mg", med.Dosage)
	assert.Equal(t, 120.50, med.Price)
	assert.Equal(t, 50, med.StockLevel)
	assert.Equal(t, uint(1), med.PharmacyID)

	// --- 2. Test GetByID ---
	idStr := strconv.Itoa(int(med.ID))
	fetchedMed, err := medicineService.GetByID(idStr)
	assert.NoError(t, err, "Fetching medicine by ID should succeed")
	assert.Equal(t, "Amoxicillin", fetchedMed.Name)

	// --- 3. Test GetAll ---
	medicines, err := medicineService.GetAll()
	assert.NoError(t, err)
	assert.Len(t, medicines, 1, "There should be exactly 1 medicine in the database")
	assert.Equal(t, "Amoxicillin", medicines[0].Name)

	// --- 4. Test Search ---
	searchResults, err := medicineService.Search("moxi")
	assert.NoError(t, err)
	assert.Len(t, searchResults, 1, "Search for 'moxi' should return Amoxicillin")
	assert.Equal(t, "Amoxicillin", searchResults[0].Name)

	// --- 5. Test UpdateMedicine ---
	// Update name and price, keep dosage same. Pass stockLevel >= 0 if we want to update it.
	// Note: UpdateMedicine might attempt to clear Redis Cache. We assume caching gracefully fails in dev.
	updatedMed, err := medicineService.UpdateMedicine(idStr, "Amoxicillin Premium", "", 150.00, 60)
	assert.NoError(t, err, "Updating medicine should succeed")
	if updatedMed != nil {
		assert.Equal(t, "Amoxicillin Premium", updatedMed.Name)
		assert.Equal(t, 150.00, updatedMed.Price)
		assert.Equal(t, "250mg", updatedMed.Dosage, "Dosage shouldn't strictly change if string is empty")
	}

	// --- 6. Test DeleteMedicine ---
	err = medicineService.DeleteMedicine(idStr)
	assert.NoError(t, err, "Deleting medicine should succeed")

	// Verify deletion
	medicinesAfterDelete, _ := medicineService.GetAll()
	assert.Len(t, medicinesAfterDelete, 0, "There should be 0 medicines after deletion")
}
