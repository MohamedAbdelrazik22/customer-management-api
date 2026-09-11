package repositories

import (
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

// GetAll retrieves every customer from the database.
func (r *CustomerRepository) GetAll() ([]models.Customer, error) {
	query := "SELECT id, name, email, status, created_at FROM customers ORDER BY id"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []models.Customer

	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}

	// rows.Err() reports any error that occurred during iteration.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return an empty slice instead of nil so JSON encodes as [] not null.
	if customers == nil {
		customers = []models.Customer{}
	}

	return customers, nil
}

// GetByID retrieves a single customer by their ID.
// Returns ErrNotFound if no such customer exists.
func (r *CustomerRepository) GetByID(id int) (*models.Customer, error) {
	query := "SELECT id, name, email, status, created_at FROM customers WHERE id = ?"

	var c models.Customer
	err := r.db.QueryRow(query, id).Scan(&c.ID, &c.Name, &c.Email, &c.Status, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &c, nil
}

// Create inserts a new customer and returns the newly created record.
func (r *CustomerRepository) Create(input models.CreateCustomerInput) (*models.Customer, error) {
	// Check for duplicate email before inserting
	if exists, err := r.emailExists(input.Email, 0); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrEmailTaken
	}

	query := "INSERT INTO customers (name, email, status) VALUES (?, ?, ?)"

	result, err := r.db.Exec(query, input.Name, input.Email, input.Status)
	if err != nil {
		return nil, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(newID))
}

// Update modifies an existing customer's fields.
// Returns ErrNotFound if the customer does not exist.
// Returns ErrEmailTaken if the email belongs to a different customer.
func (r *CustomerRepository) Update(id int, input models.UpdateCustomerInput) (*models.Customer, error) {
	// Make sure the customer exists first
	if _, err := r.GetByID(id); err != nil {
		return nil, err
	}

	// Check email uniqueness, excluding the current customer
	if exists, err := r.emailExists(input.Email, id); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrEmailTaken
	}

	query := "UPDATE customers SET name = ?, email = ?, status = ? WHERE id = ?"
	if _, err := r.db.Exec(query, input.Name, input.Email, input.Status, id); err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

// Delete removes a customer by ID.
// Returns ErrNotFound if no such customer exists.
func (r *CustomerRepository) Delete(id int) error {
	// Make sure the customer exists first
	if _, err := r.GetByID(id); err != nil {
		return err
	}

	query := "DELETE FROM customers WHERE id = ?"
	if _, err := r.db.Exec(query, id); err != nil {
		return err
	}

	return nil
}

// emailExists checks whether a given email is already used by another customer.
// Pass excludeID = 0 (or any non-positive value) when creating a new customer.
func (r *CustomerRepository) emailExists(email string, excludeID int) (bool, error) {
	var count int
	var err error

	if excludeID > 0 {
		err = r.db.QueryRow(
			"SELECT COUNT(*) FROM customers WHERE email = ? AND id != ?",
			email, excludeID,
		).Scan(&count)
	} else {
		err = r.db.QueryRow(
			"SELECT COUNT(*) FROM customers WHERE email = ?",
			email,
		).Scan(&count)
	}

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
