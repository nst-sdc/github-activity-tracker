package main

import (
	"go-orm-app/config"
	"go-orm-app/models"
	"go-orm-app/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connect DB
	config.ConnectDB()

	// Auto migrate models
	config.DB.AutoMigrate(&models.User{})

	// Setup router
	r := gin.Default()
	routes.UserRoutes(r)

	// Run server
	r.Run(":8080")
}
