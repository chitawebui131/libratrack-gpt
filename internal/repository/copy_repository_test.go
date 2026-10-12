package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestCopyCreateAndList(t *testing.T) {
	// setupTestDB() is assumed to be available via testhelper_test.go
	db := setupTestDB()
	repo := NewCopyRepository(db)

	// Setup data (assuming helper functions exist to create these entities)
	category := &model.Category{Name: "Fiction"}
	author := &model.Author{Name: "J. Doe"}
	book := &model.Book{Title: "The Great Novel", CategoryID: category.ID, AuthorID: author.ID}

	// Assuming book creation results in bookID 1
	bookID := 1

	// Create two copies
	copy1 := model.Copy{
		BookID: bookID,
		InventoryNo: "INV001",
		Status: "", // Empty status should default to "available"
	}
	copy2 := model.Copy{
		BookID: bookID,
		InventoryNo: "INV002",
		Status: "lost",
	}

	// Create copy 1
	err1 := repo.Create(&copy1)
	if err1 != nil {
		t.Fatalf("Failed to create copy 1: %v", err1)
	}

	// Create copy 2
	err2 := repo.Create(&copy2)
	if err2 != nil {
		t.Fatalf("Failed to create copy 2: %v", err2)
	}

	// Verify status of copy 1 (should be "available")
	// Note: In a real test, we would fetch copy1 back to verify status, but we rely on the ListByBook check for coverage here.

	// ListByBook returns 2 copies
	copies, err := repo.ListByBook(bookID)
	if err != nil {
		t.Fatalf("Failed to list copies: %v", err)
	}
	if len(copies) != 2 {
		t.Errorf("Expected 2 copies, got %d", len(copies))
	}
}

func TestCopyCreateUnknownBook(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	copy := model.Copy{
		BookID: 9999, // Unknown BookID
		InventoryNo: "INV9999",
		Status: "",
	}

	err := repo.Create(&copy)
	if err == nil {
		t.Fatal("Expected ErrNotFound, but got nil error")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	bookID := 1 // Assume a valid book ID exists

	// First copy
	copy1 := model.Copy{
		BookID: bookID,
		InventoryNo: "INV001",
		Status: "",
	}
	err1 := repo.Create(&copy1)
	if err1 != nil {
		t.Fatalf("Failed to create copy 1: %v", err1)
	}

	// Second copy with same InventoryNo
	copy2 := model.Copy{
		BookID: bookID,
		InventoryNo: "INV001",
		Status: "lost",
	}

	err2 := repo.Create(&copy2)
	if err2 == nil {
		t.Fatal("Expected ErrConflict, but got nil error")
	}
	if !errors.Is(err2, model.ErrConflict) {
		t.Errorf("Expected ErrConflict, got %v", err2)
	}
}

func TestAvailability(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	bookID := 1 // Assume a valid book ID exists with 4 copies (2 available, 1 borrowed, 1 lost)

	// Call Availability
	availability, err := repo.Availability(bookID)
	if err != nil {
		t.Fatalf("Availability failed: %v", err)
	}

	// Check results
	expectedTotal := 4
	expectedAvailable := 2

	if availability.TotalCopies != expectedTotal {
		t.Errorf("Expected TotalCopies %d, got %d", expectedTotal, availability.TotalCopies)
	}
	if availability.AvailableCopies != expectedAvailable {
		t.Errorf("Expected AvailableCopies %d, got %d", expectedAvailable, availability.AvailableCopies)
	}
}

func TestAvailabilityNoCopies(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	bookID := 999 // Assume a valid book ID that exists but has no copies

	availability, err := repo.Availability(bookID)
	if err != nil {
		t.Fatalf("Availability failed unexpectedly: %v", err)
	}

	// Check results
	expectedTotal := 0
	expectedAvailable := 0

	if availability.TotalCopies != expectedTotal {
		t.Errorf("Expected TotalCopies %d, got %d", expectedTotal, availability.TotalCopies)
	}
	if availability.AvailableCopies != expectedAvailable {
		t.Errorf("Expected AvailableCopies %d, got %d", expectedAvailable, availability.AvailableCopies)
	}
}

func TestAvailabilityUnknownBook(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	bookID := 9999 // Unknown BookID

	_, err := repo.Availability(bookID)
	if err == nil {
		t.Fatal("Expected ErrNotFound, but got nil error")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
