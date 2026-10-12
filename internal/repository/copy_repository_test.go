package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestCopyCreateAndList(t *testing.T) {
	// Fix: Pass *testing.T to setupTestDB
	db, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	// Setup data
	category := model.Category{Name: "Fiction"}
	author := model.Author{Name: "J. Doe"}
	book := model.Book{Title: "The Great Novel", CategoryID: category.ID, AuthorID: author.ID}

	repo := NewCopyRepository(db)

	// Create two copies
	copy1 := model.Copy{BookID: book.ID, InventoryNo: "INV001", Status: ""} // Status empty -> available
	copy2 := model.Copy{BookID: book.ID, InventoryNo: "INV002", Status: "lost"}

	_, err = repo.Create(copy1)
	if err != nil {
		t.Fatalf("Failed to create copy 1: %v", err)
	}

	_, err = repo.Create(copy2)
	if err != nil {
		t.Fatalf("Failed to create copy 2: %v", err)
	}

	// ListByBook returns 2 copies
	copies, err := repo.ListByBook(book.ID)
	if err != nil {
		t.Fatalf("Failed to list copies: %v", err)
	}
	if len(copies) != 2 {
		t.Errorf("Expected 2 copies, got %d", len(copies))
	}
}

func TestCopyCreateUnknownBook(t *testing.T) {
	db, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	repo := NewCopyRepository(db)

	// Create for BookID 9999 returns ErrNotFound.
	copy := model.Copy{BookID: 9999, InventoryNo: "INV999"}
	_, err = repo.Create(copy)

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	repo := NewCopyRepository(db)

	// Setup a book first (assuming BookID 1 exists for testing purposes)
	bookID := 1

	// Create first copy
	copy1 := model.Copy{BookID: bookID, InventoryNo: "INV001"}
	_, err = repo.Create(copy1)
	if err != nil {
		t.Fatalf("Failed to create first copy: %v", err)
	}

	// Create second copy with same InventoryNo
	copy2 := model.Copy{BookID: bookID, InventoryNo: "INV001"}
	_, err = repo.Create(copy2)

	if !errors.Is(err, ErrConflict) {
		t.Errorf("Expected ErrConflict for duplicate InventoryNo, got %v", err)
	}
}

func TestAvailability(t *testing.T) {
	db, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	repo := NewCopyRepository(db)

	// Setup book with 4 copies: 2 "available", 1 "borrowed", 1 "lost"
	bookID := 1

	// Create copies
	repo.Create(model.Copy{BookID: bookID, InventoryNo: "A1", Status: ""}) // available
	repo.Create(model.Copy{BookID: bookID, InventoryNo: "A2", Status: ""}) // available
	repo.Create(model.Copy{BookID: bookID, InventoryNo: "B1", Status: "borrowed"}) // borrowed
	repo.Create(model.Copy{BookID: bookID, InventoryNo: "L1", Status: "lost"}) // lost

	// Availability returns TotalCopies 4 and AvailableCopies 2.
	availability, err := repo.Availability(bookID)
	if err != nil {
		t.Fatalf("Failed to get availability: %v", err)
	}

	if availability.TotalCopies != 4 {
		t.Errorf("Expected TotalCopies 4, got %d", availability.TotalCopies)
	}
	if availability.AvailableCopies != 2 {
		t.Errorf("Expected AvailableCopies 2, got %d", availability.AvailableCopies)
	}
}

func TestAvailabilityNoCopies(t *testing.T) {
	db, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	repo := NewCopyRepository(db)

	// A book without copies returns 0 and 0 (not an error).
	// Assuming BookID 999 exists but has no copies.
	availability, err := repo.Availability(999)

	if err != nil {
		t.Fatalf("Expected no error for book with no copies, got %v", err)
	}

	if availability.TotalCopies != 0 || availability.AvailableCopies != 0 {
		t.Errorf("Expected TotalCopies 0 and AvailableCopies 0, got %d and %d", availability.TotalCopies, availability.AvailableCopies)
	}
}

func TestAvailabilityUnknownBook(t *testing.T) {
	db, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}
	repo := NewCopyRepository(db)

	// Availability(9999) returns ErrNotFound.
	_, err = repo.Availability(9999)

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound for unknown book, got %v", err)
	}
}
