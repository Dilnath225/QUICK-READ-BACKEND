package main

import (
	"fmt"
	"log"

	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
)

func Main() {
	config.ConnectDB()
	var users []models.User
	if err := config.DB.Order("id desc").Limit(3).Find(&users).Error; err != nil {
		log.Fatal(err)
	}

	for _, user := range users {
		fmt.Printf("ID: %d | Name: %s | Email: %s | Role: %s | Verified: %t\n",
			user.ID, user.Name, user.Email, user.Role, user.IsVerified)
	}
}
