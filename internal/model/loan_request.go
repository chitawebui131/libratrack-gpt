package model

type CreateLoanRequest struct {
	CopyID uint `json:"copy_id" binding:"required,gt=0"`
	UserID uint `json:"user_id" binding:"required,gt=0"`
}

const (
	LoanStatusActive   = "active"
	LoanStatusReturned = "returned"
	LoanStatusOverdue  = "overdue"
)

// IsValidLoanStatus reports whether s is a known loan status filter.
func IsValidLoanStatus(s string) bool {
	switch s {
	case LoanStatusActive, LoanStatusReturned, LoanStatusOverdue:
		return true
	}
	return false
}
