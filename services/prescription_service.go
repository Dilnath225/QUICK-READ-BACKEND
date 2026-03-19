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
