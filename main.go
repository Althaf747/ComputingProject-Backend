package main

import (
	"comproBackend/config"
	"comproBackend/routes"
	"comproBackend/services"
	"comproBackend/utils"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize database
	config.ConnectDatabase()

	// Initialize Firebase for push notifications
	if err := services.InitFirebase(); err != nil {
		log.Printf("Warning: Failed to initialize Firebase: %v", err)
	}

	// Start WebSocket manager for broadcasting events
	go utils.Manager.Start()

	// Create Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))

	// Setup routes
	routes.SetupRoutes(r)

	// Start server
	if err := r.Run("192.168.18.8:8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
