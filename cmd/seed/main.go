package main

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"log"
)

func main() {
	// 1. Connect to DB
	config.ConnectDB()

	// 2. Define Medicines
	medicines := []models.Medicine{
		{Name: "Amoxicillin", Dosage: "500mg", Price: 15.00, StockLevel: 100, PharmacyID: 1},
		{Name: "Amoxicillin", Dosage: "250mg", Price: 10.00, StockLevel: 100, PharmacyID: 1},
		{Name: "Panadol", Dosage: "500mg", Price: 2.50, StockLevel: 500, PharmacyID: 1},
		{Name: "Metformin", Dosage: "500mg", Price: 5.00, StockLevel: 200, PharmacyID: 1},
		{Name: "Metformin", Dosage: "850mg", Price: 7.00, StockLevel: 200, PharmacyID: 1},
		{Name: "Atorvastatin", Dosage: "10mg", Price: 12.00, StockLevel: 150, PharmacyID: 1},
		{Name: "Atorvastatin", Dosage: "20mg", Price: 18.00, StockLevel: 150, PharmacyID: 1},
		{Name: "Omeprazole", Dosage: "20mg", Price: 8.00, StockLevel: 300, PharmacyID: 1},
		{Name: "Losartan", Dosage: "50mg", Price: 6.00, StockLevel: 250, PharmacyID: 1},
		{Name: "Amlodipine", Dosage: "5mg", Price: 4.00, StockLevel: 300, PharmacyID: 1},
		{Name: "Amlodipine", Dosage: "10mg", Price: 6.00, StockLevel: 300, PharmacyID: 1},
		{Name: "Ibuprofen", Dosage: "400mg", Price: 3.50, StockLevel: 400, PharmacyID: 1},
		{Name: "Paracetamol", Dosage: "500mg", Price: 1.00, StockLevel: 1000, PharmacyID: 1},
	}

	// 3. Seed Data
	log.Println("Seeding Medicines...")
	for _, med := range medicines {
		// potential optimization: check if exists first, but for now just clear or append? 
		// simple append for "seed" is risky if run multiple times.
		// Let's use FirstOrCreate equivalent or just Create.
		var count int64
		config.DB.Model(&models.Medicine{}).Where("name = ? AND dosage = ?", med.Name, med.Dosage).Count(&count)
		if count == 0 {
			if err := config.DB.Create(&med).Error; err != nil {
				log.Printf("Failed to seed %s: %v\n", med.Name, err)
			} else {
				log.Printf("Seeded: %s %s\n", med.Name, med.Dosage)
			}
		} else {
			log.Printf("Skipped (Already Exists): %s %s\n", med.Name, med.Dosage)
		}
	}

	log.Println("Seeding Completed Successfully!")
}
