package controllers

import (
	"QUICK-READ-SYSTEM/models"
	"QUICK-READ-SYSTEM/services"
	"QUICK-READ-SYSTEM/utils"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var prescriptionService = new(services.PrescriptionService)

// POST /upload-prescription
func UploadPrescriptionImage(c *gin.Context) {
	// 1. Get the file from the request
	file, err := c.FormFile("file") // The frontend must send form-data with key "file"
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No file uploaded")
		return
	}

	// 2. Save the file locally
	// Use UUID for unique filenames
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// 3. Process with Service
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	prescription, analysis, err := prescriptionService.ProcessPrescription(currentUser.ID, filePath)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. Return Standard Response
	utils.SuccessResponse(c, "Uploaded successfully", gin.H{
		"data":     prescription,
		"analysis": analysis,
	})
}
