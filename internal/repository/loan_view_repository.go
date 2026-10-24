package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

// loanViewSelect is the shared SELECT ... FROM ... JOIN part of the loan views.
const loanViewSelect = `SELECT l.id AS loan_id, l.copy_id, c.book_id, l.user_id, c.inventory_no,
    b.title AS book_title, u.email AS user_email, u.full_name AS user_full_name,
    l.borrowed_at, l.due_at, l.returned_at, l.fine_amount
FROM loans l
JOIN copies c ON c.id = l.copy_id
JOIN books b ON b.id = c.book_id
JOIN users u ON u.id = l.user_id`

type LoanViewRepository interface {
	ListViewsByUser(userID uint, page, limit int) ([]model.LoanView, int64, error)
	ListActiveViews(page, limit int) ([]model.LoanView, int64, error)
}

type gormLoanViewRepository struct {
	db *gorm.DB
}

func NewLoanViewRepository(db *gorm.DB) LoanViewRepository {
	return &gormLoanViewRepository{db: db}
}
