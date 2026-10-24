package repository

import (
	"libratrack/internal/model"
	"time"

	"gorm.io/gorm"
)

type LoanRepository interface {
	Issue(copyID, userID uint, borrowedAt, dueAt time.Time, maxActiveLoans int) (*model.Loan, error)
	Return(loanID uint, returnedAt time.Time, calcFine func(dueAt, returnedAt time.Time) int) (*model.Loan, error)
	List(status string, page, limit int) ([]model.Loan, error)
	ListByUser(userID uint, page, limit int) ([]model.Loan, error)
}

type gormLoanRepository struct {
	db *gorm.DB
}

// Return updates the loan record with the return timestamp and fine amount.
/* func (r *gormLoanRepository) Return(loanID uint, returnedAt time.Time, calcFine func(dueAt, returnedAt time.Time) int) (*model.Loan, error) {
	var loan model.Loan
	err := r.db.First(&loan, loanID).Error
	if err != nil {
		return nil, err
	}

	// Calculate fine amount
	fineAmount := calcFine(loan.DueAt, returnedAt)
	loan.ReturnedAt = &returnedAt
	loan.FineAmount = fineAmount

	err = r.db.Save(&loan).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

// List returns a list of loans based on the status (active or returned).
func (r *gormLoanRepository) List(status string, page, limit int) ([]model.Loan, error) {
	query := r.db.Preload("Copy").Preload("User")
	switch status {
	case "active":
		query = query.Where("returned_at IS NULL")
	case "returned":
		query = query.Where("returned_at IS NOT NULL")
	}
	offset := (page - 1) * limit
	var loans []model.Loan
	err := query.Offset(offset).Limit(limit).Find(&loans).Error
	if err != nil {
		return nil, err
	}
	return loans, nil
}

// ListByUser returns a list of loans for a specific user.
func (r *gormLoanRepository) ListByUser(userID uint, page, limit int) ([]model.Loan, error) {
	offset := (page - 1) * limit
	var loans []model.Loan
	err := r.db.Preload("Copy").Preload("User").
		Where("user_id = ?", userID).
		Offset(offset).Limit(limit).
		Find(&loans).Error
	if err != nil {
		return nil, err
	}
	return loans, nil
}
*/
// ErrTooManyActiveLoans is returned when a user tries to borrow more books than allowed.
type ErrTooManyActiveLoans struct {
	UserID uint
}

func (e *ErrTooManyActiveLoans) Error() string {
	return "user has too many active loans"
}
