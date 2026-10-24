package repository

import (
	"strings"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

type BookFilter struct {
	Query         string
	CategoryID    uint
	AuthorID      uint
	OnlyAvailable bool
	Page          int
	Limit         int
}

type CatalogRepository interface {
	Search(f BookFilter) ([]model.Book, int64, error)
	GetBook(id uint) (*model.Book, error)
	AvailabilityForBooks(bookIDs []uint) (map[uint]model.CatalogAvailability, error)
}

type gormCatalogRepository struct {
	db *gorm.DB
}

// likePattern escapes the LIKE wildcards of q and wraps it in %...%.
func likePattern(q string) string {
	// 1. Replace "\" with "\\", then "%" with "\%", then "_" with "\_" (strings.ReplaceAll, in this order).
	// 2. Return "%" + escaped + "%".
	escaped := strings.ReplaceAll(q, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "%", "\\%")
	escaped = strings.ReplaceAll(escaped, "_", "\\_")
	return "%" + escaped + "%"
}

// applyBookFilter adds the WHERE conditions of f to db. All conditions are combined with AND.
func applyBookFilter(db *gorm.DB, f BookFilter) *gorm.DB {
	// 1. q := strings.TrimSpace(f.Query). If q != "": p := likePattern(q);
	//      db = db.Where("books.title ILIKE ? OR books.isbn ILIKE ? OR EXISTS (SELECT 1 FROM book_authors ba JOIN authors a ON a.id = ba.author_id WHERE ba.book_id = books.id AND a.full_name ILIKE ?)", p, p, p).
	// 2. If f.CategoryID != 0: db = db.Where("books.category_id = ?", f.CategoryID).
	// 3. If f.AuthorID != 0: db = db.Where("EXISTS (SELECT 1 FROM book_authors WHERE book_id = books.id AND author_id = ?)", f.AuthorID).
	// 4. If f.OnlyAvailable: db = db.Where("EXISTS (SELECT 1 FROM copies WHERE copies.book_id = books.id AND copies.status = 'available')").
	// 5. Return db.
	q := strings.TrimSpace(f.Query)
	if q != "" {
		p := likePattern(q)
		db = db.Where("books.title ILIKE ? OR books.isbn ILIKE ? OR EXISTS (SELECT 1 FROM book_authors ba JOIN authors a ON a.id = ba.author_id WHERE ba.book_id = books.id AND a.full_name ILIKE ?)", p, p, p)
	}
	if f.CategoryID != 0 {
		db = db.Where("books.category_id = ?", f.CategoryID)
	}
	if f.AuthorID != 0 {
		db = db.Where("EXISTS (SELECT 1 FROM book_authors WHERE book_id = books.id AND author_id = ?)", f.AuthorID)
	}
	if f.OnlyAvailable {
		db = db.Where("EXISTS (SELECT 1 FROM copies WHERE copies.book_id = books.id AND copies.status = 'available')")
	}
	return db
}
