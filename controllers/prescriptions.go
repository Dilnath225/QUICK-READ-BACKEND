package controllers

import (
	"QUICK-READ-BACKEND/models"
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
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

// GET /prescriptions
func GetUserPrescriptions(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	prescriptions, err := prescriptionService.GetUserPrescriptions(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Prescriptions retrieved", prescriptions)
}

// GET /prescriptions/:id
func GetPrescription(c *gin.Context) {
	id := c.Param("id")

	prescription, err := prescriptionService.GetPrescription(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, "Prescription details", prescription)
}

// PUT /prescriptions/:id/verify
func VerifyPrescriptionByPharmacist(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		VerifiedText string `json:"verified_text"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	prescription, err := prescriptionService.VerifyPrescription(id, input.VerifiedText)
	if err != nil {
		if err.Error() == "prescription not found" {
			utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.SuccessResponse(c, "Verified by Pharmacist", prescription)
}
