package models

import "time"

// Customer represents the customer entity in the database.
type Customer struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateCustomerInput holds the fields required to create a new customer.
type CreateCustomerInput struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

// UpdateCustomerInput holds the fields that can be updated on a customer.
type UpdateCustomerInput struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}
