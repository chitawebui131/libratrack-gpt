package handler

import (
	"bytes"
	"encoding/json"
	"errors"
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
	listFn   func(categoryID uint, page, limit int) ([]model.Book, error)
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

func (f *fakeBookRepo) List(categoryID uint, page, limit int) ([]model.Book, error) {
	if f.listFn != nil {
		return f.listFn(categoryID, page, limit)
	}
	return []model.Book{}, nil
}

func (f *fakeBookRepo) Update(book *model.Book, authorIDs []uint) error {
	return nil
}

func (f *fakeBookRepo) GetCategoryByName(name string) (*model.Category, error) {
	return nil, nil
}

// errorResponse mirrors the error envelope produced by respondError.
type errorResponse struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func newTestRouter(repo repository.BookRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewBookHandler(repo)
	r.POST("/books", h.Create)
	r.GET("/books/:id", h.GetByID)
	r.DELETE("/books/:id", h.Delete)
	r.GET("/books", h.List)
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

	var response errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if response.Error.Code != "validation_error" {
		t.Errorf("Expected error code 'validation_error', got '%s'", response.Error.Code)
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

	var response errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if response.Error.Code != "conflict" {
		t.Errorf("Expected error code 'conflict', got '%s'", response.Error.Code)
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

	var response errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if response.Error.Code != "not_found" {
		t.Errorf("Expected error code 'not_found', got '%s'", response.Error.Code)
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

	var response errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if response.Error.Code != "invalid_id" {
		t.Errorf("Expected error code 'invalid_id', got '%s'", response.Error.Code)
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

func TestListBooks200(t *testing.T) {
	repo := &fakeBookRepo{
		listFn: func(categoryID uint, page, limit int) ([]model.Book, error) {
			return []model.Book{
				{ID: 1, Title: "Book One", ISBN: "1111111111"},
				{ID: 2, Title: "Book Two", ISBN: "2222222222"},
			}, nil
		},
	}
	r := newTestRouter(repo)

	req, _ := http.NewRequest("GET", "/books?page=1&limit=10", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var response []model.Book
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 books in response, got %d", len(response))
	}
}

func TestListBooksServerError500(t *testing.T) {
	repo := &fakeBookRepo{
		listFn: func(categoryID uint, page, limit int) ([]model.Book, error) {
			return nil, errors.New("database connection failed")
		},
	}
	r := newTestRouter(repo)

	req, _ := http.NewRequest("GET", "/books?page=1&limit=10", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code 500, got %d", w.Code)
	}
}
