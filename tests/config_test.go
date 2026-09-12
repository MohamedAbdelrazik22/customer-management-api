package tests

import (
	"os"
	"testing"

	"customer-management-api/main_test_helpers"
)

// TestValidateConfig_MissingJWTSecret verifies that validateConfig returns an error
// when JWT_SECRET is not set in the environment.
// This ensures the application will not start without a configured secret.
func TestValidateConfig_MissingJWTSecret(t *testing.T) {
	// Ensure JWT_SECRET is not set
	os.Unsetenv("JWT_SECRET")

	err := main_test_helpers.ValidateConfig()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing, got nil")
	}

	expected := "JWT_SECRET is required"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

// TestValidateConfig_EmptyJWTSecret verifies that validateConfig returns an error
// when JWT_SECRET is set but empty — an empty secret is treated as missing.
func TestValidateConfig_EmptyJWTSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "")
	defer os.Unsetenv("JWT_SECRET")

	err := main_test_helpers.ValidateConfig()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is empty, got nil")
	}
}

// TestValidateConfig_WithValidSecret verifies that validateConfig succeeds
// when JWT_SECRET is set to a non-empty value.
func TestValidateConfig_WithValidSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	err := main_test_helpers.ValidateConfig()
	if err != nil {
		t.Errorf("expected no error with valid JWT_SECRET, got: %v", err)
	}
}
