package handlers

// validation_test.go tests the unexported validateCustomerInput helper within package 'handlers'.

import "testing"

func TestValidateCustomerInput_AllValid(t *testing.T) {
	err := validateCustomerInput("Mohamed Abdelrazik", "Mohamed@example.com", "active")
	if err != "" {
		t.Errorf("expected no error, got: %s", err)
	}
}

func TestValidateCustomerInput_EmptyName(t *testing.T) {
	err := validateCustomerInput("", "Mohamed@example.com", "active")
	if err != "Name is required" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestValidateCustomerInput_NameTooLong(t *testing.T) {
	longName := "a" + string(make([]byte, 100)) // 101 bytes
	err := validateCustomerInput(longName, "Mohamed@example.com", "active")
	if err != "Name must not exceed 100 characters" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestValidateCustomerInput_EmptyEmail(t *testing.T) {
	err := validateCustomerInput("Mohamed", "", "active")
	if err != "Email is required" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestValidateCustomerInput_BadEmailFormat(t *testing.T) {
	cases := []string{"notanemail", "missing@", "@nodomain.com", "no spaces@example.com"}
	for _, email := range cases {
		err := validateCustomerInput("Mohamed", email, "active")
		if err != "Invalid email format" {
			t.Errorf("email %q: expected 'Invalid email format', got: %s", email, err)
		}
	}
}

func TestValidateCustomerInput_EmptyStatus(t *testing.T) {
	err := validateCustomerInput("Mohamed", "Mohamed@example.com", "")
	if err != "Status is required" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestValidateCustomerInput_BadStatus(t *testing.T) {
	for _, status := range []string{"banned", "pending", "Active", "INACTIVE"} {
		err := validateCustomerInput("Mohamed", "Mohamed@example.com", status)
		if err != "Status must be 'active' or 'inactive'" {
			t.Errorf("status %q: unexpected error: %s", status, err)
		}
	}
}

func TestValidateCustomerInput_InactiveIsValid(t *testing.T) {
	err := validateCustomerInput("Mohamed", "Mohamed@example.com", "inactive")
	if err != "" {
		t.Errorf("expected no error for 'inactive', got: %s", err)
	}
}
