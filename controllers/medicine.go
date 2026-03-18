package controllers

import (
	"QUICK-READ-BACKEND/services"
	"QUICK-READ-BACKEND/utils"
	"net/http"
	"strconv"

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

// DELETE /medicines/:id
func DeleteMedicine(c *gin.Context) {
	id := c.Param("id")

	if err := medicineService.DeleteMedicine(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, "Medicine deleted", nil)
}

// GET /medicines Browse all medicines
func GetMedicines(c *gin.Context) {
	medicines, err := medicineService.GetAll()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, "Medicines retrieved", medicines)
}

// GET  search medicines by name
func SearchMedicines(c *gin.Context) {
	query := c.Query("q")
	medicines, err := medicineService.Search(query)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, "Search results", medicines)
}

// GET /medicines/pharmacies — Find pharmacies for medicines
func GetPharmaciesForPrescription(c *gin.Context) {
	meds := c.QueryArray("meds")
	if len(meds) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No medicines specified")
		return
	}

	pharmacies, err := medicineService.GetPharmaciesForPrescription(meds)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Pharmacies found", pharmacies)
}

// GET /medicines/price-list — Get prices for a pharmacy
func GetPharmacyPriceList(c *gin.Context) {
	pharmacyID := c.Query("pharmacy_id")
	meds := c.QueryArray("meds")

	if pharmacyID == "" || len(meds) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Pharmacy ID and medicines required")
		return
	}

	var pID uint
	id, _ := strconv.Atoi(pharmacyID)
	pID = uint(id)

	medicines, total, err := medicineService.GetPriceList(pID, meds)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, "Price list retrieved", gin.H{
		"medicines":    medicines,
		"subtotal":     total,
		"delivery_fee": 200.0, // Hardcoded for demo as per UI image
		"total":        total + 200.0,
	})
}
