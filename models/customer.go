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

// ListParams holds optional query parameters for listing customers.
type ListParams struct {
	Page   int    // 1-based page number (default 1)
	Limit  int    // rows per page (default 10, max 100)
	Search string // searches name and email (empty = no filter)
}

// PaginatedResult wraps a page of customers with metadata.
type PaginatedResult struct {
	Data       []Customer `json:"data"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	Total      int        `json:"total"`
	TotalPages int        `json:"total_pages"`
}
