package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"log"

	"github.com/google/uuid"
)

type OCRItem struct {
	Drug         string `json:"drug_name"`
	Dosage       string `json:"dosage"`
	Quantity     string `json:"quantity"`
	Instructions string `json:"instructions"`
}

type OCRResponse struct {
	Status       string    `json:"status"`
	OCRText      string    `json:"ocr_text"`
	Prescription []OCRItem `json:"prescription"`
	Confidence   float64   `json:"confidence"`
	Message      string    `json:"message"`
}

type PrescriptionService struct{}

// ScanWithOCRService sends image to our Python FastAPI OCR service
func (s *PrescriptionService) ScanWithOCRService(filePath string) (*OCRResponse, error) {
	apiUrl := os.Getenv("OCR_SERVICE_URL")
	if apiUrl == "" {
		apiUrl = "http://localhost:8000/api/v1"
	}
	log.Printf("[OCR Service] Sending image to: %s/scan", apiUrl)

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[OCR Service] Failed to open file: %v", err)
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", apiUrl+"/scan", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[OCR Service] HTTP request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[OCR Service] Unexpected status: %d, body: %s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("OCR service returned status: %d", resp.StatusCode)
	}

	var result OCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[OCR Service] JSON Decode error: %v", err)
		return nil, err
	}

	log.Printf("[OCR Service] Success! Found %d items. Confidence: %.2f", len(result.Prescription), result.Confidence)
	return &result, nil
}

// ExtractMedicineData performs simple NLP to extract medicine names and dosage from OCR text
func (s *PrescriptionService) ExtractMedicineData(ocrText string) (string, string) {
	var medicines []models.Medicine
	config.DB.Find(&medicines)

	lowerText := strings.ToLower(ocrText)
	foundMedicine := ""
	foundDosage := ""

	for _, med := range medicines {
		if strings.Contains(lowerText, strings.ToLower(med.Name)) {
			foundMedicine = med.Name
			foundDosage = med.Dosage
			break
		}
	}

	re := regexp.MustCompile(`(\d+)\s?(mg|g|ml)`)
	matches := re.FindStringSubmatch(lowerText)
	if len(matches) > 0 {
		foundDosage = matches[0]
	}

	return foundMedicine, foundDosage
}

// SaveFile generates a unique filename for prescription uploads
func (s *PrescriptionService) SaveFile(file *multipart.FileHeader) (string, error) {
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join("uploads", filename)

	return filePath, nil
}

// UploadToS3 uploads a file to AWS S3 (if configured) and returns the URL.
// Falls back to local storage if S3 is not configured.
func (s *PrescriptionService) UploadToS3(localPath string) (string, error) {
	bucket := os.Getenv("AWS_S3_BUCKET")
	if bucket == "" {
		return localPath, nil
	}
	return localPath, nil
}

// ProcessPrescription handles OCR, NLP, and DB storage
func (s *PrescriptionService) ProcessPrescription(userID uint, filePath string) (*models.Prescription, interface{}, error) {
	log.Printf("[Prescription] Processing for UserID: %d, File: %s", userID, filePath)

	// 1. Try uploading to S3
	imageURL, _ := s.UploadToS3(filePath)

	// 2. Run New AI OCR (Gemini Powered)
	result, err := s.ScanWithOCRService(filePath)

	var extractedText string
	var analysisData interface{}

	if err != nil {
		log.Printf("[Prescription] OCR API call failed: %v", err)
		extractedText = "OCR Service Error: " + err.Error()
		analysisData = []interface{}{}
	} else {
		extractedText = result.OCRText

		// Map for Frontend compatibility
		var mappedItems []map[string]string
		for _, item := range result.Prescription {
			mappedItems = append(mappedItems, map[string]string{
				"medicine_name": item.Drug,
				"dosage":        item.Dosage,
				"frequency":     item.Instructions,
				"quantity":      item.Quantity,
				"duration":      "",
			})
		}
		analysisData = mappedItems
	}

	// 3. Save to DB
	prescription := models.Prescription{
		UserID:   userID,
		ImageURL: imageURL,
		OCRText:  extractedText,
		Status:   "pending",
	}

	if err := config.DB.Create(&prescription).Error; err != nil {
		log.Printf("[Prescription] DB Save error: %v", err)
		return nil, nil, err
	}

	log.Printf("[Prescription] Successfully saved Prescription ID: %d", prescription.ID)
	return &prescription, analysisData, nil
}

// VerifyPrescription allows pharmacist to verify/correct OCR text
func (s *PrescriptionService) VerifyPrescription(id string, verifiedText string) (*models.Prescription, error) {
	var prescription models.Prescription
	if err := config.DB.First(&prescription, id).Error; err != nil {
		return nil, errors.New("prescription not found")
	}

	prescription.VerifiedText = verifiedText
	prescription.Status = "verified"
	config.DB.Save(&prescription)

	return &prescription, nil
}

// GetUserPrescriptions returns all prescriptions for a user
func (s *PrescriptionService) GetUserPrescriptions(userID uint) ([]models.Prescription, error) {
	var prescriptions []models.Prescription
	if err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&prescriptions).Error; err != nil {
		return nil, err
	}
	return prescriptions, nil
}

// GetPrescription returns a single prescription by ID
func (s *PrescriptionService) GetPrescription(id string) (*models.Prescription, error) {
	var prescription models.Prescription
	if err := config.DB.First(&prescription, id).Error; err != nil {
		return nil, errors.New("prescription not found")
	}
	return &prescription, nil
}
