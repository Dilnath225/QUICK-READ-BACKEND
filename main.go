package main

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/controllers"
	"QUICK-READ-BACKEND/pkg"

	"github.com/gin-gonic/gin"
)

func main() {
	pkg.InitLogger()
	config.ConnectDB()

	r := gin.Default()

	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.PUT("/users/profile", controllers.UpdateProfile)

	r.Run(":8080")
}
