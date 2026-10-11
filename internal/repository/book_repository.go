package repository

import (
	"errors"
	"sort"
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

// List returns a paginated list of books filtered by category.
func (r *inMemoryBookRepository) List(category string, page, limit int) ([]*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Collect copies of all stored books, applying category filter if needed.
	var all []*model.Book
	for _, book := range r.books {
		if category == "" || book.Category == category {
			copy := *book
			all = append(all, &copy)
		}
	}

	// Sort by ID ascending.
	sort.Slice(all, func(i, j int) bool {
		return all[i].ID < all[j].ID
	})

	// Pagination defaults.
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	start := (page - 1) * limit
	if start >= len(all) {
		return []*model.Book{}, nil
	}

	end := start + limit
	if end > len(all) {
		end = len(all)
	}

	return all[start:end], nil
}

// Update modifies an existing book.
func (r *inMemoryBookRepository) Update(book *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored, ok := r.books[book.ID]
	if !ok {
		return ErrNotFound
	}

	// Preserve the original CreatedAt, update UpdatedAt.
	book.CreatedAt = stored.CreatedAt
	book.UpdatedAt = time.Now()

	// Store a copy of the updated book.
	copy := *book
	r.books[book.ID] = &copy

	return nil
}

// Delete removes a book by its ID.
func (r *inMemoryBookRepository) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.books[id]; !ok {
		return ErrNotFound
	}

	delete(r.books, id)
	return nil
}
