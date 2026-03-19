package services

import (
	"context"
	"os"

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
