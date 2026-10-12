package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestCopyCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCopyRepository(db)

	// Setup data (Category, Author, Book)
	category := model.Category{Name: "Fiction"}
	author := model.Author{Name: "J. Doe"}
	book := model.Book{Title: "The Great Novel", CategoryID: category.ID, AuthorID: author.ID}

	// Create two copies
	copy1 := model.Copy{
		BookID: book.ID,
		InventoryNo: "INV001",
		Status: "", // Should default to "available"
	}
	copy2 := model.Copy{
		BookID: book.ID,
		InventoryNo: "INV002",
		Status: "lost",
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
	db := setupTestDB(t)
	repo := NewCopyRepository(db)

	copy := model.Copy{
		BookID: 9999, // Unknown BookID
		InventoryNo: "INV9999",
		Status: "",
	}

	_, err := repo.Create(copy)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCopyRepository(db)

	book := model.Book{ID: 1} // Assume Book ID 1 exists

	copy1 := model.Copy{
		BookID: book.ID,
		InventoryNo: "INV001",
		Status: "",
	}
	copy2 := model.Copy{
		BookID: book.ID,
		InventoryNo: "INV001", // Duplicate InventoryNo
		Status: "",
	}

	// Create first copy
	_, err1 := repo.Create(copy1)
	if err1 != nil {
		t.Fatalf("Failed to create copy 1: %v", err1)
	}

	// Create second copy (should fail)
	_, err2 := repo.Create(copy2)
	if !errors.Is(err2, ErrConflict) {
		t.Errorf("Expected ErrConflict, got %v", err2)
	}
}

func TestAvailability(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCopyRepository(db)

	bookID := 1 // Assume Book ID 1 exists

	// Setup 4 copies: 2 "available", 1 "borrowed", 1 "lost"
	copies := []model.Copy{
		{BookID: bookID, InventoryNo: "A1", Status: "available"},
		{BookID: bookID, InventoryNo: "A2", Status: "available"},
		{BookID: bookID, InventoryNo: "B1", Status: "borrowed"},
		{BookID: bookID, InventoryNo: "L1", Status: "lost"},
	}

	// Create copies
	for _, copy := range copies {
		_, err := repo.Create(copy)
		if err != nil {
			t.Fatalf("Failed to create copy: %v", err)
		}
	}

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
	db := setupTestDB(t)
	repo := NewCopyRepository(db)

	bookID := 99 // Assume Book ID 99 exists but has no copies

	// Check availability (should return 0, 0, no error)
	availability, err := repo.Availability(bookID)
	if err != nil {
		t.Fatalf("Availability failed unexpectedly: %v", err)
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
	repo := NewCopyRepository(db)

	bookID := 9999 // Unknown BookID

	_, err := repo.Availability(bookID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
