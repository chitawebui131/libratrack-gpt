package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

// Assuming setupTestDB and NewCopyRepository are provided by testhelper_test.go
// Assuming ErrNotFound and ErrConflict are defined in the repository package or accessible.

func TestCopyCreateAndList(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	// Setup data
	category := model.Category{Name: "Fiction"}
	author := model.Author{FirstName: "Jane", LastName: "Doe"} // Assuming FirstName/LastName fix
	book := model.Book{Title: "The Great Novel", Author: author, Category: category}

	// Create book first (assuming this is necessary for copy creation)
	// We assume a function exists to create the book in the test setup
	// For simplicity, we assume the book is already in the DB or we create it here.
	// Since the test focuses on copies, we assume the book exists.

	// Create two copies
	copy1 := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV001",
		Status:      "", // Empty status
	}
	copy2 := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV002",
		Status:      "lost",
	}

	// Create copies
	_, err1 := repo.Create(copy1)
	if err1 != nil {
		t.Fatalf("Failed to create copy 1: %v", err1)
	}

	_, err2 := repo.Create(copy2)
	if err2 != nil {
		t.Fatalf("Failed to create copy 2: %v", err2)
	}

	// Check status of copy 1 (should be "available" automatically)
	// We assume a GetCopy function exists or we check the list result.
	// For this test, we rely on ListByBook returning the correct count and status.

	// List by book
	copies, err := repo.ListByBook(book.ID)
	if err != nil {
		t.Fatalf("Failed to list copies: %v", err)
	}

	if len(copies) != 2 {
		t.Errorf("Expected 2 copies, got %d", len(copies))
	}

	// Check status of the first copy (should be "available")
	for _, copy := range copies {
		if copy.InventoryNo == "INV001" && copy.Status != "available" {
			t.Errorf("Expected copy INV001 status to be 'available', got '%s'", copy.Status)
		}
	}
}

func TestCopyCreateUnknownBook(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	copy := model.Copy{
		BookID:      9999, // Unknown BookID
		InventoryNo: "INV999",
		Status:      "",
	}

	_, err := repo.Create(copy)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound for unknown book, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	// Setup a valid book ID (assuming 1 exists)
	bookID := uint(1)

	// First copy
	copy1 := model.Copy{
		BookID:      bookID,
		InventoryNo: "INV001",
		Status:      "",
	}
	_, err1 := repo.Create(copy1)
	if err1 != nil {
		t.Fatalf("Failed to create copy 1: %v", err1)
	}

	// Second copy with same InventoryNo
	copy2 := model.Copy{
		BookID:      bookID,
		InventoryNo: "INV001",
		Status:      "lost",
	}

	_, err2 := repo.Create(copy2)
	if !errors.Is(err2, ErrConflict) {
		t.Errorf("Expected ErrConflict for duplicate inventory number, got %v", err2)
	}
}

func TestAvailability(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	// Setup book with 4 copies: 2 "available", 1 "borrowed", 1 "lost"
	bookID := uint(1)

	// Create copies
	repo.Create(model.Copy{BookID: bookID, InventoryNo: "A1", Status: "available"})
	repo.Create(model.Copy{BookID: bookID, InventoryNo: "A2", Status: "available"})
	repo.Create(model.Copy{BookID: bookID, InventoryNo: "B1", Status: "borrowed"})
	repo.Create(model.Copy{BookID: bookID, InventoryNo: "L1", Status: "lost"})

	// Check availability
	availability, err := repo.Availability(bookID)
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
	repo := NewCopyRepository(db)

	// Book ID that exists but has no copies (assuming 2 exists but has no copies)
	bookID := uint(2)

	// Check availability
	availability, err := repo.Availability(bookID)
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
	repo := NewCopyRepository(db)

	// Unknown BookID
	bookID := uint(9999)

	_, err := repo.Availability(bookID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound for unknown book, got %v", err)
	}
}
