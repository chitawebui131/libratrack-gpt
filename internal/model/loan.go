package model

import (
	"time"

	"gorm.io/gorm"
)

// Loan represents a borrowing transaction.
type Loan struct {
	gorm.Model
	CopyID     uint       `gorm:"not null" json:"copy_id"`
	UserID     uint       `gorm:"not null" json:"user_id"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	DueAt      time.Time  `json:"due_at"`
	ReturnedAt *time.Time `json:"returned_at"` // Nullable
	FineAmount int        `json:"fine_amount"` // Money is stored as an integer (UAH) and never as float
	Copy       *Copy      `json:"copy,omitempty" gorm:"foreignKey:CopyID"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// CreateLoanRequest holds data required to initiate a loan.
type CreateLoanRequest struct {
	CopyID uint `json:"copy_id" binding:"required,gt=0"`
	UserID uint `json:"user_id" binding:"required,gt=0"`
}

// LoanStatus filter constants
const (
	LoanStatusActive   = "active"
	LoanStatusReturned = "returned"
	LoanStatusOverdue  = "overdue"
)

// IsValidLoanStatus checks if the provided status string is valid.
func IsValidLoanStatus(s string) bool {
	return s == LoanStatusActive || s == LoanStatusReturned || s == LoanStatusOverdue
}
