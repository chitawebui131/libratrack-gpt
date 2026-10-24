package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
)

// Return handles POST /loans/:id/return.
func (h *LoanHandler) Return(c *gin.Context) {
	// 1. Parse the id from c.Param("id"). If it is invalid, respond like book_handler.go does for an invalid id and return.
	id, ok := parseID(c)
	if !ok {
		return
	}

	// 2. loan, err := h.svc.Return(id).
	loan, err := h.svc.Return(id)
	if err != nil {
		// 3. Map errors with errors.Is and return after responding:
		//      repository.ErrNotFound         -> 404
		//      repository.ErrAlreadyReturned  -> 409, code "already_returned"
		//      any other error                -> 500
		switch {
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		case errors.Is(err, repository.ErrAlreadyReturned):
			c.JSON(http.StatusConflict, gin.H{"code": "already_returned", "message": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}
	}

	// 4. c.JSON(http.StatusOK, loan).
	c.JSON(http.StatusOK, loan)
}
