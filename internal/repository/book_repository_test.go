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

// TestUpdate verifies that updating a book changes its title and updates UpdatedAt.
func TestUpdate(t *testing.T) {
	repo := NewInMemoryBookRepository()
	book := newTestBook("Original Title", "Fiction")
	if err := repo.Create(book); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	originalCreatedAt := book.CreatedAt

	book.Title = "Updated Title"
	if err := repo.Update(book); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	updated, err := repo.FindByID(book.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Fatalf("expected title %s, got %s", "Updated Title", updated.Title)
	}

	if updated.UpdatedAt.Before(originalCreatedAt) {
		t.Fatalf("UpdatedAt is before CreatedAt")
	}
}

// TestUpdateNotFound verifies that updating a non-existent book returns ErrNotFound.
func TestUpdateNotFound(t *testing.T) {
	repo := NewInMemoryBookRepository()
	book := &model.Book{
		ID:       999,
		Title:    "Nonexistent",
		ISBN:     "isbn-non",
		Author:   "Author",
		Category: "Fiction",
	}
	if err := repo.Update(book); err == nil {
		t.Fatalf("expected error, got nil")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestDelete verifies that deleting a book removes it from the repository.
func TestDelete(t *testing.T) {
	repo := NewInMemoryBookRepository()
	book := newTestBook("To Delete", "Fiction")
	if err := repo.Create(book); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := repo.Delete(book.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	_, err := repo.FindByID(book.ID)
	if err == nil {
		t.Fatalf("expected error after delete, got nil")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

// TestDeleteNotFound verifies that deleting a non-existent book returns ErrNotFound.
func TestDeleteNotFound(t *testing.T) {
	repo := NewInMemoryBookRepository()
	if err := repo.Delete(999); err == nil {
		t.Fatalf("expected error, got nil")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
