package controllers

import (
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

var medicineService = new(services.MedicineService)

// POST /medicines — Pharmacist adds a medicine to stock
func AddMedicine(c *gin.Context) {
	var input struct {
		Name       string  `json:"name" binding:"required"`
		Dosage     string  `json:"dosage"`
		Price      float64 `json:"price" binding:"required"`
		StockLevel int     `json:"stock_level"`
		PharmacyID uint    `json:"pharmacy_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	pharmacyID := input.PharmacyID
	if pharmacyID == 0 {
		pharmacyID = 1 // Default pharmacy
	}

	medicine, err := medicineService.AddMedicine(input.Name, input.Dosage, input.Price, input.StockLevel, pharmacyID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Medicine added", medicine)
}

// PUT /medicines/:id
func UpdateMedicine(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Name       string  `json:"name"`
		Dosage     string  `json:"dosage"`
		Price      float64 `json:"price"`
		StockLevel int     `json:"stock_level"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	medicine, err := medicineService.UpdateMedicine(id, input.Name, input.Dosage, input.Price, input.StockLevel)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, "Medicine updated", medicine)
}
