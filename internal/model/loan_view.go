package model

import "time"

// LoanView is a loan joined with its copy, book and user. It is read with raw SQL.
type LoanView struct {
	LoanID       uint       `gorm:"column:loan_id"`
	CopyID       uint       `gorm:"column:copy_id"`
	BookID       uint       `gorm:"column:book_id"`
	UserID       uint       `gorm:"column:user_id"`
	InventoryNo  string     `gorm:"column:inventory_no"`
	BookTitle    string     `gorm:"column:book_title"`
	UserEmail    string     `gorm:"column:user_email"`
	UserFullName string     `gorm:"column:user_full_name"`
	BorrowedAt   time.Time  `gorm:"column:borrowed_at"`
	DueAt        time.Time  `gorm:"column:due_at"`
	ReturnedAt   *time.Time `gorm:"column:returned_at"`
	FineAmount   int        `gorm:"column:fine_amount"`
}

func (v LoanView) IsActive() bool { return v.ReturnedAt == nil }

func (v LoanView) IsOverdue(now time.Time) bool {
	return v.ReturnedAt == nil && v.DueAt.Before(now)
}

type TopBook struct {
	BookID    uint   `gorm:"column:book_id"`
	Title     string `gorm:"column:title"`
	LoanCount int64  `gorm:"column:loan_count"`
}

// CatalogAvailability is the number of copies of one book.
type CatalogAvailability struct {
	Total     int
	Available int
}
