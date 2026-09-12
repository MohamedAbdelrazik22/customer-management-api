// Package main_test_helpers exposes functions from the main package
// that need to be tested but cannot be imported directly (since main cannot be imported).
package main_test_helpers

import (
	"errors"
	"os"
)

// ValidateConfig checks that all required environment variables are set.
// It returns an error if any required variable is missing or empty.
// This is the same logic used in main() at startup.
func ValidateConfig() error {
	if os.Getenv("JWT_SECRET") == "" {
		return errors.New("JWT_SECRET is required")
	}
	return nil
}
