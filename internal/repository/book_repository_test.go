package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

// newTestBook creates a book with the given title and category.
func newTestBook(title string, category string) *model.Book {
	return &model.Book{
		Title:    title,
		ISBN:     "isbn-" + title,
		Author:   "Author",
		Category: category,
	}
}

// TestCreateAndFindByID verifies that a book can be created and retrieved.
func TestCreateAndFindByID(t *testing.T) {
	repo := NewInMemoryBookRepository()
	book := newTestBook("Test Book", "Fiction")

	if err := repo.Create(book); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if book.ID != 1 {
		t.Fatalf("expected ID 1, got %d", book.ID)
	}

	if book.CreatedAt.IsZero() {
		t.Fatalf("CreatedAt is zero")
	}

	found, err := repo.FindByID(1)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}

	if found.Title != book.Title {
		t.Fatalf("expected title %s, got %s", book.Title, found.Title)
	}
}

// TestFindByIDNotFound verifies that FindByID returns ErrNotFound for a missing book.
func TestFindByIDNotFound(t *testing.T) {
	repo := NewInMemoryBookRepository()

	_, err := repo.FindByID(999)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
