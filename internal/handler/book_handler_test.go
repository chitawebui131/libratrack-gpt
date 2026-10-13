package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"libratrack/internal/model"
	"libratrack/internal/repository"

	"github.com/gin-gonic/gin"
)

type fakeBookRepo struct {
	createFn func(book *model.Book, authorIDs []uint) error
	findFn   func(id uint) (*model.Book, error)
	deleteFn func(id uint) error
}

func (f *fakeBookRepo) Create(book *model.Book, authorIDs []uint) error {
	return f.createFn(book, authorIDs)
}

func (f *fakeBookRepo) FindByID(id uint) (*model.Book, error) {
	return f.findFn(id)
}

func (f *fakeBookRepo) Delete(id uint) error {
	return f.deleteFn(id)
}


func (f *fakeBookRepo) List(page uint, limit int, _ int) ([]model.Book, error) {
	return []model.Book{}, nil
}

func (f *fakeBookRepo) Update(book *model.Book, authorIDs []uint) error {
	return nil
}

func (f *fakeBookRepo) GetCategoryByName(name string) (*model.Category, error) {
	return nil, nil
}


func newTestRouter(repo repository.BookRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewBookHandler(repo)
	r.POST("/books", h.Create)
	r.GET("/books/:id", h.GetByID)
	r.DELETE("/books/:id", h.Delete)
	return r
}

func TestCreateBook201(t *testing.T) {
	repo := &fakeBookRepo{
		createFn: func(book *model.Book, authorIDs []uint) error {
			book.ID = 1
			return nil
		},
	}
	r := newTestRouter(repo)

	body := bytes.NewBufferString(`{"title":"T","isbn":"1234567890","category_id":1,"author_ids":[1]}`)
	req, _ := http.NewRequest("POST", "/books", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d", w.Code)
	}

	var response struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected book ID 1, got %d", response.ID)
	}
}

func TestCreateBookValidation422(t *testing.T) {
	repo := &fakeBookRepo{}
	r := newTestRouter(repo)

	body := bytes.NewBufferString(`{"isbn":"1234567890"}`)
	req, _ := http.NewRequest("POST", "/books", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status code 422, got %d", w.Code)
	}

	var response struct {
		Code string `json:"error.code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if response.Code != "validation_error" {
		t.Errorf("Expected error code 'validation_error', got '%s'", response.Code)
	}
}

func TestCreateBookConflict409(t *testing.T) {
	repo := &fakeBookRepo{
		createFn: func(book *model.Book, authorIDs []uint) error {
			return repository.ErrConflict
		},
	}
	r := newTestRouter(repo)

	body := bytes.NewBufferString(`{"title":"T","isbn":"1234567890","category_id":1,"author_ids":[1]}`)
	req, _ := http.NewRequest("POST", "/books", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status code 409, got %d", w.Code)
	}

	var response struct {
		Code string `json:"error.code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if response.Code != "conflict" {
		t.Errorf("Expected error code 'conflict', got '%s'", response.Code)
	}
}

func TestGetBookNotFound404(t *testing.T) {
	repo := &fakeBookRepo{
		findFn: func(id uint) (*model.Book, error) {
			return nil, repository.ErrNotFound
		},
	}
	r := newTestRouter(repo)

	req, _ := http.NewRequest("GET", "/books/1", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", w.Code)
	}

	var response struct {
		Code string `json:"error.code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if response.Code != "not_found" {
		t.Errorf("Expected error code 'not_found', got '%s'", response.Code)
	}
}

func TestGetBookInvalidID400(t *testing.T) {
	repo := &fakeBookRepo{}
	r := newTestRouter(repo)

	req, _ := http.NewRequest("GET", "/books/abc", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}

	var response struct {
		Code string `json:"error.code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if response.Code != "invalid_id" {
		t.Errorf("Expected error code 'invalid_id', got '%s'", response.Code)
	}
}

func TestDeleteBook204(t *testing.T) {
	repo := &fakeBookRepo{
		deleteFn: func(id uint) error {
			return nil
		},
	}
	r := newTestRouter(repo)

	req, _ := http.NewRequest("DELETE", "/books/1", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status code 204, got %d", w.Code)
	}
}
