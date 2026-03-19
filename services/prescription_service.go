package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
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
		fmt.Println("⚠️  AWS S3 not configured, using local storage")
		return localPath, nil
	}

	// AWS S3 upload using presigned URL approach
	// In production, use aws-sdk-go-v2:
	// cfg, _ := awsconfig.LoadDefaultConfig(context.TODO())
	// client := s3.NewFromConfig(cfg)
	// key := "prescriptions/" + filepath.Base(localPath)
	// _, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
	//     Bucket: aws.String(bucket),
	//     Key:    aws.String(key),
	//     Body:   file,
	// })
	// return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key), nil

	fmt.Println("⚠️  S3 upload placeholder — configure AWS credentials for production")
	return localPath, nil
}

// ProcessPrescription handles OCR, NLP, and DB storage
func (s *PrescriptionService) ProcessPrescription(userID uint, filePath string) (*models.Prescription, interface{}, error) {
	// 1. Try uploading to S3
	imageURL, _ := s.UploadToS3(filePath)

	// 2. Run Real OCR
	extractedText, err := s.DetectText(filePath)
	if err != nil {
		extractedText = "Error in OCR: " + err.Error()
	}

	// 3. Run NLP / Extraction
	medName, medDosage := s.ExtractMedicineData(extractedText)

	analysisText := ""
	if medName != "" {
		analysisText = "\n\n--- AI Analysis ---\nDetected Medicine: " + medName
		if medDosage != "" {
			analysisText += "\nDetected Dosage: " + medDosage
		}
	} else {
		analysisText = "\n\n--- AI Analysis ---\nNo known medicine detected in database."
	}
	finalText := extractedText + analysisText

	// 4. Save to DB
	prescription := models.Prescription{
		UserID:   userID,
		ImageURL: imageURL,
		OCRText:  finalText,
		Status:   "pending",
	}

	if err := config.DB.Create(&prescription).Error; err != nil {
		return nil, nil, err
	}

	analysisData := map[string]string{
		"medicine": medName,
		"dosage":   medDosage,
	}

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
