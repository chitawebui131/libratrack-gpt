package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

// List returns loans, newest first. status is optional: "" means all.
func (r *gormLoanRepository) List(status string, page, limit int) ([]model.Loan, error) {
	// 1. offset, size := pagingBounds(page, limit).
	offset, size := pagingBounds(page, limit)
	// 2. q := r.db.Model(&model.Loan{}).
	q := r.db.Model(&model.Loan{})
	// 3. switch status:
	//      "active"   -> q = q.Where("returned_at IS NULL")
	//      "returned" -> q = q.Where("returned_at IS NOT NULL")
	//      "overdue"  -> q = q.Where("returned_at IS NULL AND due_at < now()")
	switch status {
	case "active":
		q = q.Where("returned_at IS NULL")
	case "returned":
		q = q.Where("returned_at IS NOT NULL")
	case "overdue":
		q = q.Where("returned_at IS NULL AND due_at < now()")
	}
	// 4. var loans []model.Loan.
	var loans []model.Loan
	// 5. err := q.Preload("Copy.Book").Preload("User").Order("borrowed_at DESC, id DESC").Offset(offset).Limit(size).Find(&loans).Error.
	err := q.Preload("Copy.Book").Preload("User").Order("borrowed_at DESC, id DESC").Offset(offset).Limit(size).Find(&loans).Error
	// 6. If err != nil, return nil, translateLoanError(err). Otherwise return loans, nil.
	if err != nil {
		return nil, translateLoanError(err)
	}
	return loans, nil
}

// ListByUser returns the loans of one user, newest first.
func (r *gormLoanRepository) ListByUser(userID uint, page, limit int) ([]model.Loan, error) {
	// 1. offset, size := pagingBounds(page, limit).
	offset, size := pagingBounds(page, limit)
	// 2. var loans []model.Loan.
	var loans []model.Loan
	// 3. err := r.db.Where("user_id = ?", userID).Preload("Copy.Book").Order("borrowed_at DESC, id DESC").Offset(offset).Limit(size).Find(&loans).Error.
	err := r.db.Where("user_id = ?", userID).Preload("Copy.Book").Order("borrowed_at DESC, id DESC").Offset(offset).Limit(size).Find(&loans).Error
	// 4. If err != nil, return nil, translateLoanError(err). Otherwise return loans, nil.
	if err != nil {
		return nil, translateLoanError(err)
	}
	return loans, nil
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	// Return &gormLoanRepository{db: db}.
	return &gormLoanRepository{db: db}
}
