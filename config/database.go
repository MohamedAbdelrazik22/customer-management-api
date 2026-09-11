package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// ConnectDB reads environment variables, opens a MySQL connection,
// and returns the *sql.DB instance.  The caller is responsible for
// closing it when the application shuts down.
func ConnectDB() *sql.DB {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3307")
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "customer_management")

	// Build the DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Verify the connection is actually reachable
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established successfully")
	return db
}

// getEnv returns the value of an environment variable, or a default
// value if the variable is not set.
func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}
