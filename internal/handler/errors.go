package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"libratrack/internal/repository"
)

// respondError sends an error response with the given status and details.
func respondError(c *gin.Context, status int, code, message, details string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
			"details": details,
		},
	})
}

// respondValidation sends a 422 validation error response.
func respondValidation(c *gin.Context, err error) {
	respondError(c, http.StatusUnprocessableEntity, "validation_error", "request validation failed", err.Error())
}

// respondRepoError maps repository errors to HTTP responses.
func respondRepoError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		respondError(c, http.StatusNotFound, "not_found", "resource not found", "")
	case errors.Is(err, repository.ErrConflict):
		respondError(c, http.StatusConflict, "conflict", "resource already exists", "")
	case errors.Is(err, repository.ErrInvalidReference):
		respondError(c, http.StatusUnprocessableEntity, "invalid_reference", "referenced resource does not exist", "")
	default:
		respondError(c, http.StatusInternalServerError, "internal_error", "internal server error", "")
	}
}

// parseID parses the "id" path parameter and validates it.
func parseID(c *gin.Context) (uint, bool) {
	v, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || v == 0 {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid id", "")
		return 0, false
	}
	return uint(v), true
}
