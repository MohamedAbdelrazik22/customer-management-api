package main

// seed.go — run this once to create a test admin user
// Usage: go run seed/seed.go
// This file is NOT part of the main application.

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"customer-management-api/config"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using environment variables")
	}

	// Override port to match your setup
	os.Setenv("DB_PORT", "3307")
	os.Setenv("DB_PASSWORD", "MySQL@123")

	db := config.ConnectDB()
	defer db.Close()

	username := "admin"
	password := "admin123"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO users (username, password) VALUES (?, ?) ON DUPLICATE KEY UPDATE password = VALUES(password)",
		username, string(hash),
	)
	if err != nil {
		log.Fatalf("Failed to insert user: %v", err)
	}

	fmt.Printf("✅ User '%s' created successfully.\n", username)
	fmt.Printf("   Username: %s\n", username)
	fmt.Printf("   Password: %s\n", password)
}
