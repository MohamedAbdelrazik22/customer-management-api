package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"customer-management-api/config"
	"customer-management-api/handlers"
	"customer-management-api/repositories"
	"customer-management-api/routes"
)

func main() {
	// Load .env file if it exists 
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using environment variables")
	}

	// Connect to the database
	db := config.ConnectDB()
	defer db.Close()

	// Wire up the layers: repository -> handler -> router
	customerRepo := repositories.NewCustomerRepository(db)
	customerHandler := handlers.NewCustomerHandler(customerRepo)

	userRepo := repositories.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo)

	// Create the Gin engine with default middleware (Logger + Recovery)
	r := gin.Default()

	// Register all routes
	routes.SetupRoutes(r, customerHandler, authHandler)

	// Determine which port to listen on
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
