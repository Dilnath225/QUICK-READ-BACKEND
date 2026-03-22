package main

import (
	"QUICK-READ-BACKEND/Routes"
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
	config.InitLogger()
	// defer config.Logger.Sync() // Uncomment if using zap directly in main

	// 2. Connect to Database (AWS RDS / Local Postgres)
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

	// 4. Initialize Event Bus (for async processing)
	services.InitEventBus()

	// 5. Initialize WebSocket Hub
	controllers.InitWebSocketHub()

	// 6. Setup Router
	r := gin.Default()

	// CORS Configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // In production, restrict this
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup all routes from Routes package
	Routes.SetupRouter(r)

	// Serve uploads directory statically
	r.Static("/uploads", "./uploads")

	// 7. Start Server
	log.Println("🚀 Starting Quick-Read API on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
