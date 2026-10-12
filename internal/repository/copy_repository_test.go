package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

// Assuming testhelper_test.go provides setupTestDB and necessary helpers.

func TestCopyCreateAndList(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	// Setup required entities
	category := model.Category{Name: "Fiction"}
	author := model.Author{Name: "J. Doe"}
	book := model.Book{Title: "The Great Novel", CategoryID: category.ID, AuthorID: author.ID}

	// Ensure entities exist in the test DB
	db.Create(&category)
	db.Create(&author)
	db.Create(&book)

	// Create two copies
	// Copy 1: Status empty (should default to available)
	copy1 := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV001",
		Status:      "", // Empty status
	}
	err := repo.Create(&copy1)
	if err != nil {
		t.Fatalf("Failed to create copy 1: %v", err)
	}

	// Copy 2: Status "lost"
	copy2 := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV002",
		Status:      "lost",
	}
	err = repo.Create(&copy2)
	if err != nil {
		t.Fatalf("Failed to create copy 2: %v", err)
	}

	// Check status of the first copy (should be "available")
	var retrievedCopy1 model.Copy
	db.First(&retrievedCopy1, copy1.ID)
	if retrievedCopy1.Status != "available" {
		t.Errorf("Expected copy 1 status to be 'available', got '%s'", retrievedCopy1.Status)
	}

	// ListByBook should return 2 copies
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
	repo := NewCopyRepository(db)

	// Try creating a copy for BookID 9999 (unknown)
	copy := model.Copy{
		BookID:      9999,
		InventoryNo: "INV9999",
		Status:      "available",
	}

	err := repo.Create(&copy)
	if err == nil {
		t.Fatal("Expected ErrNotFound when creating copy for unknown book, got nil")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	// Setup required entities
	category := model.Category{Name: "Fiction"}
	author := model.Author{Name: "J. Doe"}
	book := model.Book{Title: "The Great Novel", CategoryID: category.ID, AuthorID: author.ID}

	db.Create(&category)
	db.Create(&author)
	db.Create(&book)

	// Create first copy
	copy1 := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV001",
		Status:      "available",
	}
	err := repo.Create(&copy1)
	if err != nil {
		t.Fatalf("Failed to create copy 1: %v", err)
	}

	// Create second copy with the same InventoryNo
	copy2 := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV001", // Duplicate
		Status:      "borrowed",
	}

	err = repo.Create(&copy2)
	if err == nil {
		t.Fatal("Expected ErrConflict when creating duplicate InventoryNo, got nil")
	}
	if !errors.Is(err, model.ErrConflict) {
		t.Errorf("Expected ErrConflict, got %v", err)
	}
}

func TestAvailability(t *testing.T) {
	db := setupTestDB()
	repo := NewCopyRepository(db)

	// Setup required entities
	category := model.Category{Name: "Fiction"}
	author := model.Author{Name: "J. Doe"}
	book := model.Book{Title: "The Great Novel", CategoryID: category.ID, AuthorID: author.ID}

	db.Create(&category)
	db.Create(&author)
	db.Create(&book)

	// Create 4 copies: 2 available, 1 borrowed, 1 lost
	copies := []model.Copy{
		{BookID: book.ID, InventoryNo: "A1", Status: "available"},
		{BookID: book.ID, InventoryNo: "A2", Status: "available"},
		{BookID: book.ID, InventoryNo: "B1", Status: "borrowed"},
		{BookID: book.ID, InventoryNo: "L1", Status: "lost"},
	}

	for _, copy := range copies {
		err := repo.Create(&copy)
		if err != nil {
			t.Fatalf("Failed to create copy: %v", err)
		}
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
	repo := NewCopyRepository(db)

	// Setup a book that exists but has no copies
	category := model.Category{Name: "Fiction"}
	author := model.Author{Name: "J. Doe"}
	book := model.Book{Title: "No Copies Book", CategoryID: category.ID, AuthorID: author.ID}

	db.Create(&category)
	db.Create(&author)
	db.Create(&book)

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
	repo := NewCopyRepository(db)

	// Try checking availability for BookID 9999 (unknown)
	_, err := repo.Availability(9999)
	if err == nil {
		t.Fatal("Expected ErrNotFound when checking availability for unknown book, got nil")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
