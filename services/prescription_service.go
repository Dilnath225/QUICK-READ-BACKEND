package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"context"
	"os"
	"regexp"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
)

type PrescriptionService struct{}

// DetectText sends image to Google Cloud Vision for OCR
func (s *PrescriptionService) DetectText(filePath string) (string, error) {
	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	image, err := vision.NewImageFromReader(file)
	if err != nil {
		return "", err
	}

	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		return "", err
	}

	if len(annotations) == 0 {
		return "No text found", nil
	}

	return annotations[0].Description, nil
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
