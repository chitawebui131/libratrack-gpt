package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestCopyCreateAndList(t *testing.T) {
	db := setupTestDB()
	// Removed db.Close() defer call to fix vet error
	repo := NewCopyRepository(db)

	// Setup data
	category := model.Category{Name: "Fiction"}
	author := model.Author{Name: "J. Doe"}
	book := model.Book{Title: "The Great Novel", CategoryID: category.ID, AuthorID: author.ID}

	// Create two copies
	// Copy 1: status empty (should become "available")
	copy1 := model.Copy{
		BookID: book.ID,
		InventoryNo: "INV001",
		Status: "", // Empty status
	}
	repo.Create(&copy1)

	// Copy 2: status "lost"
	copy2 := model.Copy{
		BookID: book.ID,
		InventoryNo: "INV002",
		Status: "lost",
	}
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
	db := setupTestDB()
	// Removed db.Close() defer call to fix vet error
	repo := NewCopyRepository(db)

	// Try to create a copy for BookID 9999
	copy := model.Copy{
		BookID: 9999,
		InventoryNo: "INV9999",
		Status: "",
	}

	_, err := repo.Create(&copy)
	if err == nil {
		t.Error("Expected ErrNotFound, got nil")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := setupTestDB()
	// Removed db.Close() defer call to fix vet error
	repo := NewCopyRepository(db)

	// Setup a book
	book := model.Book{Title: "Test Book"}
	repo.Create(&book)

	// First copy
	copy1 := model.Copy{
		BookID: book.ID,
		InventoryNo: "DUPINV",
		Status: "available",
	}
	repo.Create(&copy1)

	// Second copy with same InventoryNo
	copy2 := model.Copy{
		BookID: book.ID,
		InventoryNo: "DUPINV",
		Status: "borrowed",
	}

	_, err := repo.Create(&copy2)
	if err == nil {
		t.Error("Expected ErrConflict, got nil")
	}
	if !errors.Is(err, model.ErrConflict) {
		t.Errorf("Expected ErrConflict, got %v", err)
	}
}

func TestAvailability(t *testing.T) {
	db := setupTestDB()
	// Removed db.Close() defer call to fix vet error
	repo := NewCopyRepository(db)

	// Setup book
	book := model.Book{Title: "Availability Test Book"}
	repo.Create(&book)

	// Setup 4 copies: 2 "available", 1 "borrowed", 1 "lost"
	copies := []model.Copy{
		{BookID: book.ID, InventoryNo: "A1", Status: "available"},
		{BookID: book.ID, InventoryNo: "A2", Status: "available"},
		{BookID: book.ID, InventoryNo: "B1", Status: "borrowed"},
		{BookID: book.ID, InventoryNo: "L1", Status: "lost"},
	}

	for i := range copies {
		repo.Create(&copies[i])
	}

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
	db := setupTestDB()
	// Removed db.Close() defer call to fix vet error
	repo := NewCopyRepository(db)

	// Book without copies
	book := model.Book{Title: "No Copies Book"}
	repo.Create(&book)

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
	db := setupTestDB()
	// Removed db.Close() defer call to fix vet error
	repo := NewCopyRepository(db)

	// Check availability for BookID 9999
	_, err := repo.Availability(9999)
	if err == nil {
		t.Error("Expected ErrNotFound, got nil")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
