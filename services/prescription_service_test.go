package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessPrescription(t *testing.T) {
	// 1. Mock the OCR Service
	mockOCRResponse := OCRResponse{
		Status:  "success",
		OCRText: "Amoxicillin 500mg\nTake 1 capsule every 8 hours\nQty: 21",
		Prescription: []OCRItem{
			{
				Drug:         "Amoxicillin",
				Dosage:       "500mg",
				Quantity:     "21",
				Instructions: "Take 1 capsule every 8 hours",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockOCRResponse)
	}))
	defer server.Close()

	// 2. Setup Environment
	os.Setenv("OCR_SERVICE_URL", server.URL)
	defer os.Unsetenv("OCR_SERVICE_URL")

	// 3. Setup Test Data
	config.DB.Exec("DELETE FROM prescriptions")
	config.DB.Exec("DELETE FROM users")
	
	user := models.User{Name: "Test Patient", Email: "test@patient.com"}
	config.DB.Create(&user)

	// Create a dummy file
	dummyFile := "test_prescription.jpg"
	os.WriteFile(dummyFile, []byte("dummy image data"), 0644)
	defer os.Remove(dummyFile)

	// 4. Run the test
	prescriptionService := new(PrescriptionService)
	prescription, analysis, err := prescriptionService.ProcessPrescription(user.ID, dummyFile)

	// 5. Assertions
	assert.NoError(t, err)
	assert.NotNil(t, prescription)
	assert.Equal(t, user.ID, prescription.UserID)
	
	analysisList := analysis.([]map[string]string)
	assert.Len(t, analysisList, 1)
	assert.Equal(t, "Amoxicillin", analysisList[0]["medicine_name"])
	assert.Equal(t, "500mg", analysisList[0]["dosage"])
	assert.Equal(t, "21", analysisList[0]["quantity"])
}
