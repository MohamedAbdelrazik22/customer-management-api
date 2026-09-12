package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"customer-management-api/config"
	"customer-management-api/handlers"
	"customer-management-api/repositories"
	"customer-management-api/routes"
)

// validateConfig checks that all required environment variables are set.
// It returns an error if any required variable is missing or empty.
func validateConfig() error {
	if os.Getenv("JWT_SECRET") == "" {
		return errors.New("JWT_SECRET is required")
	}
	return nil
}

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using environment variables")
	}

	// Validate required configuration before starting
	if err := validateConfig(); err != nil {
		log.Fatalf("Configuration error: %v", err)
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

	// Create the HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start the server in a goroutine so it doesn't block signal handling
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for termination signal (Ctrl+C or SIGTERM from OS/container)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server — waiting for in-flight requests to finish...")

	// Give in-flight requests up to 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}
