package model

import "time"

// Loan represents a borrowing transaction.
type Loan struct {
	ID         uint       `gorm:"primaryKey"`
	CopyID     uint       `gorm:"index;not null"`
	UserID     uint       `gorm:"index;not null"`
	BorrowedAt time.Time  `gorm:"not null"`
	DueAt      time.Time  `gorm:"not null"`
	ReturnedAt *time.Time `gorm:"column:returned_at"`    // nullable
	FineAmount int        `gorm:"type:integer;not null"` // money is stored as an integer and never as float
	Copy       *Copy      `json:"copy,omitempty" gorm:"foreignKey:CopyID"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// LoanStatus constants
const (
	LoanStatusActive   = "active"
	LoanStatusReturned = "returned"
	LoanStatusOverdue  = "overdue"
)

// IsValidLoanStatus checks if the provided loan status string is valid.
func IsValidLoanStatus(s string) bool {
	return s == LoanStatusActive || s == LoanStatusReturned || s == LoanStatusOverdue
}

// CreateLoanRequest is used to initiate a new loan.
type CreateLoanRequest struct {
	CopyID uint `json:"copy_id" binding:"required,gt=0"`
	UserID uint `json:"user_id" binding:"required,gt=0"`
}
