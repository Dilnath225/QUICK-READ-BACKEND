package main

import (
	"QUICK-READ-BACKEND/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	checkDB()
}

func checkDB() {
	err := godotenv.Load("../.env") // Adjusted path for .env since it's now in scripts/
	if err != nil {
		log.Println("Error loading .env file from scripts/ directory, trying root...")
		err = godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var users []models.User
	db.Find(&users)

	fmt.Println("--- Users in Database ---")
	for _, u := range users {
		fmt.Printf("ID: %v | Name: %s | Email: %s | Role: %s\n", u.ID, u.Name, u.Email, u.Role)
	}
}
