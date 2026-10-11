package repository

import (
	"errors"

	"libratrack/internal/model"
)

// ErrNotFound is returned when a book cannot be found.
var ErrNotFound = errors.New("book not found")

// BookRepository defines the operations for managing books.
type BookRepository interface {
	// Create inserts a new book into the repository.
	Create(book *model.Book) error

	// FindByID retrieves a book by its ID.
	FindByID(id uint) (*model.Book, error)

	// List returns a paginated list of books filtered by category.
	// If category is empty, no category filter is applied.
	// Page starts at 1.
	List(category string, page, limit int) ([]*model.Book, error)

	// Update modifies an existing book.
	Update(book *model.Book) error

	// Delete removes a book by its ID.
	Delete(id uint) error
}
