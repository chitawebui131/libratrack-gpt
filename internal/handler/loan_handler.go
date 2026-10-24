package handler

import (
	"errors"
	"net/http"

	"libratrack/internal/model"
	"libratrack/internal/repository"

	"github.com/gin-gonic/gin"
)

type LoanService interface {
	Issue(copyID, userID uint) (*model.Loan, error)
	Return(loanID uint) (*model.Loan, error)
	List(status string, page, limit int) ([]model.Loan, error)
	ListByUser(userID uint, page, limit int) ([]model.Loan, error)
}

type LoanHandler struct {
	svc LoanService
}

func NewLoanHandler(svc LoanService) *LoanHandler {
	return &LoanHandler{svc: svc}
}

// Create handles POST /loans.
func (h *LoanHandler) Create(c *gin.Context) {
	var req model.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "validation failed", "error": err.Error()})
		return
	}

	loan, err := h.svc.Issue(req.CopyID, req.UserID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidReference):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
			return
		case errors.Is(err, repository.ErrCopyNotAvailable):
			c.JSON(http.StatusConflict, gin.H{"code": "copy_not_available", "message": err.Error()})
			return
		case errors.Is(err, repository.ErrLoanLimitReached):
			c.JSON(http.StatusConflict, gin.H{"code": "loan_limit_reached", "message": err.Error()})
			return
		case errors.Is(err, repository.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}
	}

	c.JSON(http.StatusCreated, loan)
}

// ListByUser handles GET /me/loans.
func (h *LoanHandler) ListByUser(c *gin.Context) {
	// TODO: Extract userID from auth middleware context
	var userID uint = 42

	loans, err := h.svc.ListByUser(userID, 1, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, loans)
}
