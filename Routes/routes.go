package Routes

import (
	"QUICK-READ-BACKEND/controllers"
	"QUICK-READ-BACKEND/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	// Root Health Check
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":  "QUICK-READ-BACKEND API is running!",
			"api_base": "/api",
		})
	})

	api := r.Group("/api")
	{
		// Simple Health Check
		api.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Welcome to QUICK-READ-BACKEND API",
				"status":  "Running",
			})
		})

		api.POST("/register", controllers.Register) // Patients, Riders, Pharmacists
		api.POST("/verify-otp", controllers.VerifyOTP)
		api.POST("/login", controllers.Login)

		// Medicines
		api.GET("/medicines", middleware.CacheResponse(5*time.Minute), controllers.GetMedicines)
		api.GET("/medicines/search", controllers.SearchMedicines)
		api.GET("/medicines/pharmacies", controllers.GetPharmaciesForPrescription)
		api.GET("/medicines/price-list", middleware.CacheResponse(5*time.Minute), controllers.GetPharmacyPriceList)

		authorized := api.Group("/")
		authorized.Use(middleware.RequireAuth)
		{
			// Common Profile Route
			authorized.GET("/users/profile", controllers.GetProfile)
			authorized.PUT("/users/profile", controllers.UpdateProfile) // Only used by Patients right now
			authorized.POST("/users/profile-picture", controllers.UploadProfilePicture)

			//patient specific routes
			patientRoutes := authorized.Group("/")
			patientRoutes.Use(middleware.RequireRole("patient"))
			{
				// Prescriptions (Both Patients and Pharmacists can upload)
				authorized.POST("/prescriptions", middleware.RequireRole("patient", "pharmacist"), controllers.UploadPrescriptionImage)
				authorized.GET("/prescriptions", middleware.RequireRole("patient", "pharmacist"), controllers.GetUserPrescriptions)

				// Cart
				patientRoutes.POST("/cart", controllers.AddToCart)
				patientRoutes.GET("/cart", controllers.GetCart)
				patientRoutes.DELETE("/cart/:id", controllers.RemoveFromCart)
				patientRoutes.DELETE("/cart", controllers.ClearCart)

				// Orders
				patientRoutes.POST("/orders", controllers.CreateOrder)
				patientRoutes.GET("/orders", controllers.GetOrderHistory)
				patientRoutes.POST("/orders/:id/checkout", controllers.InitializeCheckout)
				patientRoutes.POST("/orders/:id/confirm-payment", controllers.ConfirmPayment)
			}

			// Shared Route: Get Prescription (Pharmacist verifying or Patient viewing)
			authorized.GET("/prescriptions/:id", middleware.RequireRole("patient", "pharmacist"), controllers.GetPrescription)

			// Shared Route: Track Order (Pharmacist verifying delivery coordinates or Patient viewing)
			authorized.GET("/orders/:id/track", middleware.RequireRole("patient", "pharmacist", "delivery", "driver"), controllers.TrackOrder)

			// PHARMACIST ROUTES

			pharmacyRoutes := authorized.Group("/pharmacy")
			pharmacyRoutes.Use(middleware.RequireRole("pharmacist"))
			{
				// Prescriptions verification
				pharmacyRoutes.PUT("/prescriptions/:id/verify", controllers.VerifyPrescriptionByPharmacist)

				// Medicine Management
				pharmacyRoutes.POST("/medicines", controllers.AddMedicine)
				pharmacyRoutes.PUT("/medicines/:id", controllers.UpdateMedicine)
				pharmacyRoutes.DELETE("/medicines/:id", controllers.DeleteMedicine)

				// Dashboard Reporting
				pharmacyRoutes.GET("/dashboard/revenue", controllers.GetDailyRevenue)
				pharmacyRoutes.GET("/dashboard/orders", controllers.GetOrderStats)
				pharmacyRoutes.GET("/orders", controllers.GetPharmacyOrders)
			}

			// Shared Order Status Update
			authorized.PUT("/orders/:id/status", middleware.RequireRole("pharmacist", "delivery", "driver"), controllers.UpdateOrderStatus)

			// DELIVERY SYSTEM ROUTES
			authorized.POST("/delivery/assign/:orderId", middleware.RequireRole("pharmacist", "system"), controllers.AssignDriver)
			authorized.GET("/delivery/route/:orderId", middleware.RequireRole("delivery", "driver", "pharmacist"), controllers.GetRoute)
			authorized.GET("/delivery/order-history/:orderId", middleware.RequireRole("patient", "pharmacist"), controllers.GetOrderDeliveryHistory)

			// Driver specific endpoints
			driverRoutes := authorized.Group("/delivery")
			driverRoutes.Use(middleware.RequireRole("delivery", "driver", "patient"))
			{
				driverRoutes.POST("/register", controllers.RegisterDriver)
				driverRoutes.GET("/profile", controllers.GetDriverProfile)
				driverRoutes.PUT("/profile", controllers.UpdateDriverProfile)
				driverRoutes.POST("/availability", controllers.ToggleAvailability)
				driverRoutes.POST("/location", controllers.UpdateDriverLocation)
				driverRoutes.GET("/history", controllers.GetDeliveryHistory)
				driverRoutes.GET("/orders", controllers.GetActiveOrders)
				driverRoutes.GET("/stats", controllers.GetDriverStats)
			}

			// Delivery Driver updating order status/location
			authorized.PUT("/orders/:id/location", middleware.RequireRole("delivery", "driver"), controllers.UpdateLocation)

			// NOTIFICATIONS

			authorized.GET("/notifications", controllers.GetNotifications)
			authorized.PUT("/notifications/:id/read", controllers.MarkNotificationRead)
			authorized.PUT("/notifications/read-all", controllers.MarkAllNotificationsRead)
		}

		//       WEBSOCKETS

		api.GET("/ws/track/:orderId", controllers.HandleWebSocket)
	}

}
