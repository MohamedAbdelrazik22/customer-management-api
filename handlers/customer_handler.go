package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"customer-management-api/models"
	"customer-management-api/repositories"
)

// emailRegex is a simple pattern used to validate email addresses.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// CustomerRepositoryInterface defines the methods the handler depends on.
// This allows tests to swap in a fake repository without a real database.
type CustomerRepositoryInterface interface {
	GetAll(params models.ListParams) (*models.PaginatedResult, error)
	GetByID(id int) (*models.Customer, error)
	Create(input models.CreateCustomerInput) (*models.Customer, error)
	Update(id int, input models.UpdateCustomerInput) (*models.Customer, error)
	Delete(id int) error
}

// CustomerHandler holds a reference to the customer repository interface.
type CustomerHandler struct {
	repo CustomerRepositoryInterface
}

// NewCustomerHandler creates a CustomerHandler backed by the real MySQL repository.
func NewCustomerHandler(repo *repositories.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

// NewCustomerHandlerWithInterface creates a CustomerHandler with any repository
// implementation — useful for unit tests with a fake repository.
func NewCustomerHandlerWithInterface(repo CustomerRepositoryInterface) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

// GetAll handles GET /customers
// Supports optional query parameters:
//
//	?page=1      (default: 1)
//	?limit=10    (default: 10, max: 100)
//	?search=ahmed (searches name and email)
func (h *CustomerHandler) GetAll(c *gin.Context) {
	params := models.ListParams{
		Page:   parseQueryInt(c, "page", 1),
		Limit:  parseQueryInt(c, "limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
	}

	result, err := h.repo.GetAll(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve customers"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID handles GET /customers/:id
// Returns a single customer or 404 if not found.
func (h *CustomerHandler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	customer, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customer})
}

// Create handles POST /customers
// Validates the request body and creates a new customer.
func (h *CustomerHandler) Create(c *gin.Context) {
	var input models.CreateCustomerInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	// Trim whitespace before validation
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Status = strings.TrimSpace(input.Status)

	if validationErr := validateCustomerInput(input.Name, input.Email, input.Status); validationErr != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		return
	}

	customer, err := h.repo.Create(input)
	if err != nil {
		if errors.Is(err, repositories.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email is already in use"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": customer})
}

// Update handles PUT /customers/:id
// Validates the request body and updates the customer.
func (h *CustomerHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input models.UpdateCustomerInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	// Trim whitespace before validation
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Status = strings.TrimSpace(input.Status)

	if validationErr := validateCustomerInput(input.Name, input.Email, input.Status); validationErr != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		return
	}

	customer, err := h.repo.Update(id, input)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
			return
		}
		if errors.Is(err, repositories.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email is already in use by another customer"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customer})
}

// Delete handles DELETE /customers/:id
// Removes the customer with the given ID.
func (h *CustomerHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	err := h.repo.Delete(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// parseID parses the :id URL parameter and writes a 400 response if invalid.
// Returns the id and true on success, or 0 and false on failure.
func parseID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return 0, false
	}
	return id, true
}

// parseQueryInt reads a query parameter as an integer.
// Falls back to defaultVal if the parameter is missing or not a valid integer.
func parseQueryInt(c *gin.Context, key string, defaultVal int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val < 1 {
		return defaultVal
	}
	return val
}

// validateCustomerInput checks name, email and status fields.
// Returns a non-empty error string on the first validation failure, or "" if valid.
func validateCustomerInput(name, email, status string) string {
	if name == "" {
		return "Name is required"
	}
	if len(name) > 100 {
		return "Name must not exceed 100 characters"
	}
	if email == "" {
		return "Email is required"
	}
	if !emailRegex.MatchString(email) {
		return "Invalid email format"
	}
	if status == "" {
		return "Status is required"
	}
	if status != "active" && status != "inactive" {
		return "Status must be 'active' or 'inactive'"
	}
	return ""
}
