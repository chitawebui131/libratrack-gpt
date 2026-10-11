package handler

import (
	"errors"
	"libratrack/internal/model"
	"libratrack/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BookHandler handles book-related HTTP requests.
type BookHandler struct {
	repo repository.BookRepository
}

// NewBookHandler creates a new BookHandler with the given repository.
func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

// ErrorBody represents the body of an error response.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

// ErrorResponse represents the full error response payload.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// respondError sends an error response with the given status and details.
func respondError(c *gin.Context, status int, code, message, details string) {
	c.AbortWithStatusJSON(status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// Create handles the creation of a new book.
func (h *BookHandler) Create(c *gin.Context) {
	var req model.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "validation_error", "invalid request body", err.Error())
		return
	}

	book := model.Book{
		Title:         req.Title,
		ISBN:          req.ISBN,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	}

	if err := h.repo.Create(book); err != nil {
		respondError(c, http.StatusInternalServerError, "internal_error", "could not create book", "")
		return
	}

	c.JSON(http.StatusCreated, book)
}

// parseID reads the "id" path parameter, validates it, and returns the parsed uint ID.
// If the ID is invalid or zero, it responds with a 422 error and returns 0, false.
func parseID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		respondError(c, http.StatusUnprocessableEntity, "validation_error", "invalid id", "")
		return 0, false
	}
	return uint(id64), true
}

// GetByID retrieves a book by its ID.
func (h *BookHandler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	book, err := h.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(c, http.StatusNotFound, "not_found", "book not found", "")
			return
		}
		respondError(c, http.StatusInternalServerError, "internal_error", "could not load book", "")
		return
	}

	c.JSON(http.StatusOK, book)
}

// List retrieves a paginated list of books, optionally filtered by category.
func (h *BookHandler) List(c *gin.Context) {
	category := c.Query("category")

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		respondError(c, http.StatusUnprocessableEntity, "validation_error", "invalid pagination", "")
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		respondError(c, http.StatusUnprocessableEntity, "validation_error", "invalid pagination", "")
		return
	}

	if limit > 100 {
		limit = 100
	}

	books, err := h.repo.List(category, page, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal_error", "could not list books", "")
		return
	}

	c.JSON(http.StatusOK, books)
}

// Update handles updating an existing book.
func (h *BookHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req model.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "validation_error", "invalid request body", err.Error())
		return
	}

	book, err := h.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(c, http.StatusNotFound, "not_found", "book not found", "")
			return
		}
		respondError(c, http.StatusInternalServerError, "internal_error", "could not load book", "")
		return
	}

	book.Title = req.Title
	book.ISBN = req.ISBN
	book.Author = req.Author
	book.Category = req.Category
	book.PublishedYear = req.PublishedYear
	book.Description = req.Description

	if err := h.repo.Update(book); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(c, http.StatusNotFound, "not_found", "book not found", "")
			return
		}
		respondError(c, http.StatusInternalServerError, "internal_error", "could not update book", "")
		return
	}

	c.JSON(http.StatusOK, book)
}
