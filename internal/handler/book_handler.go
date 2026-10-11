package handler

import (
	"libratrack/internal/repository"

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
