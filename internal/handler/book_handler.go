package handler

import (
	"libratrack/internal/model"
	"libratrack/internal/repository"
	"net/http"

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
