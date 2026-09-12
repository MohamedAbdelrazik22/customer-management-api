package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"customer-management-api/handlers"
	"customer-management-api/middleware"
	"customer-management-api/models"
	"customer-management-api/repositories"
)

// --------------------------------------------------------------------------
// Fake repository — satisfies handlers.CustomerRepositoryInterface
// No real database is needed for these tests.
// --------------------------------------------------------------------------

type fakeRepo struct {
	customers map[int]*models.Customer
	nextID    int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		customers: make(map[int]*models.Customer),
		nextID:    1,
	}
}

func (f *fakeRepo) GetAll(_ context.Context, params models.ListParams) (*models.PaginatedResult, error) {
	list := []models.Customer{}
	for _, c := range f.customers {
		list = append(list, *c)
	}
	return &models.PaginatedResult{
		Data:       list,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      len(list),
		TotalPages: 1,
	}, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id int) (*models.Customer, error) {
	c, ok := f.customers[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return c, nil
}

func (f *fakeRepo) Create(_ context.Context, input models.CreateCustomerInput) (*models.Customer, error) {
	for _, c := range f.customers {
		if c.Email == input.Email {
			return nil, repositories.ErrEmailTaken
		}
	}
	c := &models.Customer{
		ID:     f.nextID,
		Name:   input.Name,
		Email:  input.Email,
		Status: input.Status,
	}
	f.customers[f.nextID] = c
	f.nextID++
	return c, nil
}

func (f *fakeRepo) Update(_ context.Context, id int, input models.UpdateCustomerInput) (*models.Customer, error) {
	c, ok := f.customers[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	for oid, oc := range f.customers {
		if oc.Email == input.Email && oid != id {
			return nil, repositories.ErrEmailTaken
		}
	}
	c.Name = input.Name
	c.Email = input.Email
	c.Status = input.Status
	return c, nil
}

func (f *fakeRepo) Delete(_ context.Context, id int) error {
	if _, ok := f.customers[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(f.customers, id)
	return nil
}

// --------------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------------

func setupRouter(repo handlers.CustomerRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewCustomerHandlerWithInterface(repo)
	r.GET("/customers", h.GetAll)
	r.GET("/customers/:id", h.GetByID)
	r.POST("/customers", h.Create)
	r.PUT("/customers/:id", h.Update)
	r.DELETE("/customers/:id", h.Delete)
	return r
}

func toJSON(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("toJSON: %v", err)
	}
	return bytes.NewBuffer(b)
}

// --------------------------------------------------------------------------
// POST /customers — Create tests
// --------------------------------------------------------------------------

func TestCreate_Success(t *testing.T) {
	r := setupRouter(newFakeRepo())
	body := toJSON(t, map[string]string{
		"name": "Mohamed Abdelrazik", "email": "Mohamed@example.com", "status": "active",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreate_MissingName(t *testing.T) {
	r := setupRouter(newFakeRepo())
	body := toJSON(t, map[string]string{"email": "Mohamed@example.com", "status": "active"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreate_InvalidEmail(t *testing.T) {
	r := setupRouter(newFakeRepo())
	body := toJSON(t, map[string]string{
		"name": "Mohamed", "email": "not-an-email", "status": "active",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreate_InvalidStatus(t *testing.T) {
	r := setupRouter(newFakeRepo())
	body := toJSON(t, map[string]string{
		"name": "Mohamed", "email": "Mohamed@example.com", "status": "banned",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// Test 2 — Duplicate Customer Email: creating a second customer with the same email
// must return 409 Conflict with error "Email already exists".
func TestCreate_DuplicateEmail(t *testing.T) {
	repo := newFakeRepo()
	r := setupRouter(repo)

	// Create the first customer
	body := toJSON(t, map[string]string{
		"name": "Mohamed Abdelrazik", "email": "Mohamed@example.com", "status": "active",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("setup: expected 201 for first customer, got %d", w.Code)
	}

	// Try to create a second customer with the same email
	body = toJSON(t, map[string]string{
		"name": "Someone Else", "email": "Mohamed@example.com", "status": "active",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d — body: %s", w.Code, w.Body.String())
	}

	// Verify the error message
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["error"] != "Email already exists" {
		t.Errorf("expected error 'Email already exists', got %q", resp["error"])
	}
}

// --------------------------------------------------------------------------
// GET /customers/:id — GetByID tests
// --------------------------------------------------------------------------

func TestGetByID_NotFound(t *testing.T) {
	r := setupRouter(newFakeRepo())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/customers/999", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// Test 3 — Invalid Customer ID: a non-numeric ID must return 400 Bad Request.
func TestGetByID_InvalidID(t *testing.T) {
	r := setupRouter(newFakeRepo())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/customers/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["error"] != "Invalid customer ID" {
		t.Errorf("expected error 'Invalid customer ID', got %q", resp["error"])
	}
}

func TestGetByID_Found(t *testing.T) {
	repo := newFakeRepo()
	r := setupRouter(repo)

	// Create one customer first
	body := toJSON(t, map[string]string{
		"name": "Sara", "email": "sara@example.com", "status": "inactive",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Now retrieve it
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/customers/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// DELETE /customers/:id — Delete tests
func TestDelete_NotFound(t *testing.T) {
	r := setupRouter(newFakeRepo())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/customers/999", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDelete_Success(t *testing.T) {
	repo := newFakeRepo()
	r := setupRouter(repo)

	// Create a customer to delete
	body := toJSON(t, map[string]string{
		"name": "Temp", "email": "temp@example.com", "status": "active",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Delete it
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/customers/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// PUT /customers/:id — Update tests
func TestUpdate_NotFound(t *testing.T) {
	r := setupRouter(newFakeRepo())
	body := toJSON(t, map[string]string{
		"name": "X", "email": "x@example.com", "status": "active",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/customers/999", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdate_EmailConflict(t *testing.T) {
	repo := newFakeRepo()
	r := setupRouter(repo)

	// Create two customers
	for _, email := range []string{"a@example.com", "b@example.com"} {
		body := toJSON(t, map[string]string{"name": "Test", "email": email, "status": "active"})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/customers", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
	}

	// Try to update customer 2 with customer 1's email
	body := toJSON(t, map[string]string{
		"name": "Test", "email": "a@example.com", "status": "active",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/customers/2", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

// --------------------------------------------------------------------------
// Test 4 — Unauthorized Request
// Protected endpoints must return 401 when no valid JWT token is provided.
// --------------------------------------------------------------------------

// setupProtectedRouter creates a router with the real AuthMiddleware applied to POST /customers,
// so we can test the 401 behavior without a real database.
func setupProtectedRouter(repo handlers.CustomerRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewCustomerHandlerWithInterface(repo)

	// Public read routes
	r.GET("/customers", h.GetAll)
	r.GET("/customers/:id", h.GetByID)

	// Protected routes — require a valid JWT
	protected := r.Group("/customers")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("", h.Create)
		protected.PUT("/:id", h.Update)
		protected.DELETE("/:id", h.Delete)
	}
	return r
}

// TestCreate_NoToken verifies that POST /customers returns 401 when no Authorization header is provided.
func TestCreate_NoToken(t *testing.T) {
	// JWT_SECRET must be set for the middleware to work correctly
	os.Setenv("JWT_SECRET", "test-secret-for-auth-tests")
	defer os.Unsetenv("JWT_SECRET")

	r := setupProtectedRouter(newFakeRepo())
	body := toJSON(t, map[string]string{
		"name": "Test User", "email": "test@example.com", "status": "active",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header set — request should be rejected
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestCreate_InvalidToken verifies that POST /customers returns 401 when an invalid JWT is provided.
func TestCreate_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-auth-tests")
	defer os.Unsetenv("JWT_SECRET")

	r := setupProtectedRouter(newFakeRepo())
	body := toJSON(t, map[string]string{
		"name": "Test User", "email": "test@example.com", "status": "active",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer this.is.not.a.valid.token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for invalid token, got %d — body: %s", w.Code, w.Body.String())
	}
}

// TestCreate_MalformedAuthHeader verifies that a malformed Authorization header returns 401.
func TestCreate_MalformedAuthHeader(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-auth-tests")
	defer os.Unsetenv("JWT_SECRET")

	r := setupProtectedRouter(newFakeRepo())
	body := toJSON(t, map[string]string{
		"name": "Test User", "email": "test@example.com", "status": "active",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "NotBearer sometoken")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for malformed auth header, got %d", w.Code)
	}
}

// Ensure the package compiles even if errors import is unused
var _ = errors.New
