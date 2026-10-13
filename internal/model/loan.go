package model

import (
	"time"
)

// LoanStatus defines the possible states of a loan.
type LoanStatus string

const (
	LoanStatusActive    LoanStatus = "active"
	LoanStatusReturned  LoanStatus = "returned"
	LoanStatusOverdue   LoanStatus = "overdue"
)

// IsValidLoanStatus checks if the provided loan status string is valid.
func IsValidLoanStatus(s string) bool {
	return s == string(LoanStatusActive) || s == string(LoanStatusReturned) || s == string(LoanStatusOverdue)
}

// Loan represents a borrowing transaction.
type Loan struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CopyID      uint      `gorm:"not null" json:"copy_id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	BorrowedAt  time.Time `json:"borrowed_at"`
	DueAt       time.Time `json:"due_at"`
	ReturnedAt  *time.Time `gorm:"column:returned_at" json:"returned_at"` // Nullable
	FineAmount  int       `gorm:"default:0" json:"fine_amount"` // Money is stored as an integer and never as float
	Copy        *Copy     `json:"copy,omitempty" gorm:"foreignKey:CopyID"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateLoanRequest defines the input structure for creating a new loan.
type CreateLoanRequest struct {
	CopyID uint `json:"copy_id" binding:"required,gt=0"`
	UserID uint `json:"user_id" binding:"required,gt=0"`
}
