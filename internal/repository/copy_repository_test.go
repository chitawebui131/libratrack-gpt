package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestCopyCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	category := model.Category{Name: "Fiction"}
	author := model.Author{Name: "J. Doe"}
	book := model.Book{Title: "The Great Novel", CategoryID: category.ID, AuthorID: author.ID}

	repo := NewCopyRepository(db)

	// Create two copies
	// Copy 1: status empty (should default to available)
	copy1 := model.Copy{BookID: book.ID, InventoryNo: "INV001", Status: ""}
	repo.Create(&copy1)

	// Copy 2: status "lost"
	copy2 := model.Copy{BookID: book.ID, InventoryNo: "INV002", Status: "lost"}
	repo.Create(&copy2)

	// ListByBook returns 2 copies
	copies, err := repo.ListByBook(book.ID)
	if err != nil {
		t.Fatalf("ListByBook failed: %v", err)
	}
	if len(copies) != 2 {
		t.Errorf("Expected 2 copies, got %d", len(copies))
	}
}

func TestCopyCreateUnknownBook(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewCopyRepository(db)

	// BookID 9999 does not exist
	copy := model.Copy{BookID: 9999, InventoryNo: "INV9999"}

	err := repo.Create(&copy)
	if err == nil {
		t.Error("Expected ErrNotFound, got nil")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Setup book
	book := model.Book{Title: "Test Book"}
	db.Create(&book)

	repo := NewCopyRepository(db)

	// First copy
	copy1 := model.Copy{BookID: book.ID, InventoryNo: "DUP001"}
	repo.Create(&copy1)

	// Second copy with same InventoryNo
	copy2 := model.Copy{BookID: book.ID, InventoryNo: "DUP001"}

	err := repo.Create(&copy2)
	if err == nil {
		t.Error("Expected ErrConflict, got nil")
	}
	if !errors.Is(err, model.ErrConflict) {
		t.Errorf("Expected ErrConflict, got %v", err)
	}
}

func TestAvailability(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Setup book with 4 copies: 2 "available", 1 "borrowed", 1 "lost"
	book := model.Book{Title: "Availability Test Book"}
	db.Create(&book)

	repo := NewCopyRepository(db)

	// Create copies
	// Available 1
	repo.Create(&model.Copy{BookID: book.ID, InventoryNo: "A1", Status: "available"})
	// Available 2
	repo.Create(&model.Copy{BookID: book.ID, InventoryNo: "A2", Status: "available"})
	// Borrowed
	repo.Create(&model.Copy{BookID: book.ID, InventoryNo: "B1", Status: "borrowed"})
	// Lost
	repo.Create(&model.Copy{BookID: book.ID, InventoryNo: "L1", Status: "lost"})

	// Check availability
	availability, err := repo.Availability(book.ID)
	if err != nil {
		t.Fatalf("Availability failed: %v", err)
	}

	if availability.TotalCopies != 4 {
		t.Errorf("Expected TotalCopies 4, got %d", availability.TotalCopies)
	}
	if availability.AvailableCopies != 2 {
		t.Errorf("Expected AvailableCopies 2, got %d", availability.AvailableCopies)
	}
}

func TestAvailabilityNoCopies(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Book without copies
	book := model.Book{Title: "No Copies Book"}
	db.Create(&book)

	repo := NewCopyRepository(db)

	// Check availability
	availability, err := repo.Availability(book.ID)
	if err != nil {
		t.Fatalf("Availability failed: %v", err)
	}

	if availability.TotalCopies != 0 {
		t.Errorf("Expected TotalCopies 0, got %d", availability.TotalCopies)
	}
	if availability.AvailableCopies != 0 {
		t.Errorf("Expected AvailableCopies 0, got %d", availability.AvailableCopies)
	}
}

func TestAvailabilityUnknownBook(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewCopyRepository(db)

	// BookID 9999 does not exist
	_, err := repo.Availability(9999)
	if err == nil {
		t.Error("Expected ErrNotFound, got nil")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
