package main

import (
	"QUICK-READ-BACKEND/Routes" // Still in root for now, but referenced
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/controllers"
	"QUICK-READ-BACKEND/models"
	"QUICK-READ-BACKEND/services"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Initialize Logger
	// config.InitLogger() // Using pkg/logger instead in professional setup

	// 2. Connect to Database
	config.ConnectDB()

	// AutoMigrate all models
	err := config.DB.AutoMigrate(
		&models.User{},
		&models.Prescription{},
		&models.Medicine{},
		&models.Order{},
		&models.Cart{},
		&models.CartItem{},
		&models.DeliveryDriver{},
		&models.DeliveryHistory{},
		&models.Route{},
		&models.Notification{},
	)
	if err != nil {
		log.Fatalf("❌ Failed to migrate database models: %v", err)
	}

	// Ensure uploads directory exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	// 3. Initialize Redis Cache
	config.ConnectRedis()

	// 4. Initialize Event Bus
	services.InitEventBus()

	// 5. Initialize WebSocket Hub
	controllers.InitWebSocketHub()

	// 6. Setup Router
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	Routes.SetupRouter(r)

	log.Println("🚀 Starting Quick-Read API on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
