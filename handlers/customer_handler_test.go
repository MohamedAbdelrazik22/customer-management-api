package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"customer-management-api/handlers"
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

func (f *fakeRepo) GetAll(params models.ListParams) (*models.PaginatedResult, error) {
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

func (f *fakeRepo) GetByID(id int) (*models.Customer, error) {
	c, ok := f.customers[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return c, nil
}

func (f *fakeRepo) Create(input models.CreateCustomerInput) (*models.Customer, error) {
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

func (f *fakeRepo) Update(id int, input models.UpdateCustomerInput) (*models.Customer, error) {
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

func (f *fakeRepo) Delete(id int) error {
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

	// Try to create a second customer with the same email
	body = toJSON(t, map[string]string{
		"name": "Someone Else", "email": "Mohamed@example.com", "status": "active",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/customers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
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

func TestGetByID_InvalidID(t *testing.T) {
	r := setupRouter(newFakeRepo())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/customers/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
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

// Ensure the package compiles even if errors import is unused
var _ = errors.New
