package repository

import (
	"errors"
	"sync"
	"time"

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

// inMemoryBookRepository is an in-memory implementation of BookRepository.
type inMemoryBookRepository struct {
	mu     sync.RWMutex
	books  map[uint]*model.Book
	nextID uint
}

// Create inserts a new book into the repository.
func (r *inMemoryBookRepository) Create(book *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	book.ID = r.nextID
	r.nextID++

	now := time.Now()
	book.CreatedAt = now
	book.UpdatedAt = now

	// Store a copy of the book to avoid external mutation.
	copy := *book
	r.books[book.ID] = &copy

	return nil
}

// FindByID retrieves a book by its ID.
func (r *inMemoryBookRepository) FindByID(id uint) (*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stored, ok := r.books[id]
	if !ok {
		return nil, ErrNotFound
	}

	// Return a copy to prevent external mutation.
	copy := *stored
	return &copy, nil
}
