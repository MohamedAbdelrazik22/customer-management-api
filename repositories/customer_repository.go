package repositories

import (
	"context"
	"database/sql"
	"errors"

	"customer-management-api/models"
)

// ErrNotFound is returned when a customer does not exist in the database.
var ErrNotFound = errors.New("customer not found")

// ErrEmailTaken is returned when an email is already in use by another customer.
var ErrEmailTaken = errors.New("email already in use")

// CustomerRepository handles all database operations for customers.
type CustomerRepository struct {
	db *sql.DB
}

// NewCustomerRepository creates a new CustomerRepository with the given database.
func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// GetAll retrieves customers with optional search and pagination.
// params.Search filters by name or email (case-insensitive LIKE match).
// params.Page and params.Limit control which page of results is returned.
func (r *CustomerRepository) GetAll(ctx context.Context, params models.ListParams) (*models.PaginatedResult, error) {
	// Sanitize pagination values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	// Build WHERE clause for optional search
	var args []interface{}
	whereClause := ""
	if params.Search != "" {
		whereClause = " WHERE name LIKE ? OR email LIKE ?"
		like := "%" + params.Search + "%"
		args = append(args, like, like)
	}

	// Count total matching rows (for pagination metadata)
	var total int
	countQuery := "SELECT COUNT(*) FROM customers" + whereClause
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Fetch the requested page
	offset := (params.Page - 1) * params.Limit
	dataQuery := "SELECT id, name, email, status, created_at FROM customers" +
		whereClause + " ORDER BY id LIMIT ? OFFSET ?"
	args = append(args, params.Limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := []models.Customer{}
	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Calculate total pages (at least 1)
	totalPages := total / params.Limit
	if total%params.Limit != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	return &models.PaginatedResult{
		Data:       customers,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetByID retrieves a single customer by their ID.
// Returns ErrNotFound if no such customer exists.
func (r *CustomerRepository) GetByID(ctx context.Context, id int) (*models.Customer, error) {
	query := "SELECT id, name, email, status, created_at FROM customers WHERE id = ?"

	var c models.Customer
	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Name, &c.Email, &c.Status, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &c, nil
}

// Create inserts a new customer and returns the newly created record.
func (r *CustomerRepository) Create(ctx context.Context, input models.CreateCustomerInput) (*models.Customer, error) {
	// Check for duplicate email before inserting
	if exists, err := r.emailExists(ctx, input.Email, 0); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrEmailTaken
	}

	query := "INSERT INTO customers (name, email, status) VALUES (?, ?, ?)"

	result, err := r.db.ExecContext(ctx, query, input.Name, input.Email, input.Status)
	if err != nil {
		return nil, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, int(newID))
}

// Update modifies an existing customer's fields.
// Returns ErrNotFound if the customer does not exist.
// Returns ErrEmailTaken if the email belongs to a different customer.
func (r *CustomerRepository) Update(ctx context.Context, id int, input models.UpdateCustomerInput) (*models.Customer, error) {
	// Make sure the customer exists first
	if _, err := r.GetByID(ctx, id); err != nil {
		return nil, err
	}

	// Check email uniqueness, excluding the current customer
	if exists, err := r.emailExists(ctx, input.Email, id); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrEmailTaken
	}

	query := "UPDATE customers SET name = ?, email = ?, status = ? WHERE id = ?"
	if _, err := r.db.ExecContext(ctx, query, input.Name, input.Email, input.Status, id); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

// Delete removes a customer by ID.
// Returns ErrNotFound if no such customer exists.
func (r *CustomerRepository) Delete(ctx context.Context, id int) error {
	// Make sure the customer exists first
	if _, err := r.GetByID(ctx, id); err != nil {
		return err
	}

	query := "DELETE FROM customers WHERE id = ?"
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return err
	}

	return nil
}

// emailExists checks whether a given email is already used by another customer.
// Pass excludeID = 0 (or any non-positive value) when creating a new customer.
func (r *CustomerRepository) emailExists(ctx context.Context, email string, excludeID int) (bool, error) {
	var count int
	var err error

	if excludeID > 0 {
		err = r.db.QueryRowContext(
			ctx,
			"SELECT COUNT(*) FROM customers WHERE email = ? AND id != ?",
			email, excludeID,
		).Scan(&count)
	} else {
		err = r.db.QueryRowContext(
			ctx,
			"SELECT COUNT(*) FROM customers WHERE email = ?",
			email,
		).Scan(&count)
	}

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
