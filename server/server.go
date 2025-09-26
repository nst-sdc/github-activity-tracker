package main

import (
	"log"

	"github.com/nst-sdc/github-activity-tracker/database"
	"github.com/nst-sdc/github-activity-tracker/handler"
	"github.com/nst-sdc/github-activity-tracker/service"
	"github.com/nst-sdc/github-activity-tracker/store"

	"gofr.dev/pkg/gofr"
)

func main() {
	app := gofr.New()

	var userStore store.UserStore

	// Try to initialize database, fallback to in-memory if it fails
	if err := database.InitDatabase(); err != nil {
		log.Printf("Database connection failed, using in-memory store: %v", err)
		userStore = store.NewInMemoryStore()
	} else {
		log.Println("Using PostgreSQL database store")
		userStore = store.NewPostgreSQLStore(database.GetDB())
	}

	// Dependencies
	userService := service.NewUserService(userStore)
	userHandler := handler.NewUserHandler(*userService)

	// Routes
	app.POST("/user", userHandler.AddGitId)

	// Health check endpoint
	app.GET("/health", func(ctx *gofr.Context) (interface{}, error) {
		return map[string]string{
			"status":   "healthy",
			"database": "connected",
		}, nil
	})

	log.Println("Server starting with PostgreSQL database...")

	// Start server
	app.Run()
}
