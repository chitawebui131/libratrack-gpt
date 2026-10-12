package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestCopyCreateAndList(t *testing.T) {
	db := setupTestDB(t)

	// Create a category
	cat := createCategory(t, db, "Fiction")

	// Create an author
	auth := createAuthor(t, db, "John", "Doe")

	// Create a book
	book := createBook(t, db, "Test Book", cat.ID, auth.ID)

	repo := NewCopyRepository(db)

	// Create first copy with empty status
	copy1 := &model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV001",
		Status:      "",
	}
	if err := repo.Create(copy1); err != nil {
		t.Fatalf("unexpected error creating copy1: %v", err)
	}
	if copy1.Status != "available" {
		t.Fatalf("expected status 'available', got %s", copy1.Status)
	}

	// Create second copy with status "lost"
	copy2 := &model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV002",
		Status:      "lost",
	}
	if err := repo.Create(copy2); err != nil {
		t.Fatalf("unexpected error creating copy2: %v", err)
	}

	copies, err := repo.ListByBook(book.ID)
	if err != nil {
		t.Fatalf("unexpected error listing copies: %v", err)
	}
	if len(copies) != 2 {
		t.Fatalf("expected 2 copies, got %d", len(copies))
	}
}

func TestCopyCreateUnknownBook(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCopyRepository(db)

	copy := &model.Copy{
		BookID:      9999,
		InventoryNo: "INV999",
		Status:      "",
	}
	err := repo.Create(copy)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := setupTestDB(t)

	// Create a category
	cat := createCategory(t, db, "Fiction")

	// Create an author
	auth := createAuthor(t, db, "Jane", "Smith")

	// Create a book
	book := createBook(t, db, "Duplicate Test", cat.ID, auth.ID)

	repo := NewCopyRepository(db)

	copy1 := &model.Copy{
		BookID:      book.ID,
		InventoryNo: "DUPLICATE",
		Status:      "",
	}
	if err := repo.Create(copy1); err != nil {
		t.Fatalf("unexpected error creating first copy: %v", err)
	}

	copy2 := &model.Copy{
		BookID:      book.ID,
		InventoryNo: "DUPLICATE",
		Status:      "",
	}
	err := repo.Create(copy2)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestAvailability(t *testing.T) {
	db := setupTestDB(t)

	// Create a category
	cat := createCategory(t, db, "Non-Fiction")

	// Create an author
	auth := createAuthor(t, db, "Alice", "Wonder")

	// Create a book
	book := createBook(t, db, "Availability Test", cat.ID, auth.ID)

	repo := NewCopyRepository(db)

	// Create copies with various statuses
	copies := []model.Copy{
		{BookID: book.ID, InventoryNo: "A1", Status: "available"},
		{BookID: book.ID, InventoryNo: "A2", Status: "available"},
		{BookID: book.ID, InventoryNo: "B1", Status: "borrowed"},
		{BookID: book.ID, InventoryNo: "L1", Status: "lost"},
	}
	for i := range copies {
		if err := repo.Create(&copies[i]); err != nil {
			t.Fatalf("unexpected error creating copy %d: %v", i, err)
		}
	}

	avail, err := repo.Availability(book.ID)
	if err != nil {
		t.Fatalf("unexpected error getting availability: %v", err)
	}
	if avail.TotalCopies != 4 {
		t.Fatalf("expected TotalCopies 4, got %d", avail.TotalCopies)
	}
	if avail.AvailableCopies != 2 {
		t.Fatalf("expected AvailableCopies 2, got %d", avail.AvailableCopies)
	}
}

func TestAvailabilityNoCopies(t *testing.T) {
	db := setupTestDB(t)

	// Create a category
	cat := createCategory(t, db, "Empty")

	// Create an author
	auth := createAuthor(t, db, "Bob", "Empty")

	// Create a book with no copies
	book := createBook(t, db, "No Copies", cat.ID, auth.ID)

	repo := NewCopyRepository(db)

	avail, err := repo.Availability(book.ID)
	if err != nil {
		t.Fatalf("unexpected error getting availability: %v", err)
	}
	if avail.TotalCopies != 0 {
		t.Fatalf("expected TotalCopies 0, got %d", avail.TotalCopies)
	}
	if avail.AvailableCopies != 0 {
		t.Fatalf("expected AvailableCopies 0, got %d", avail.AvailableCopies)
	}
}

func TestAvailabilityUnknownBook(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCopyRepository(db)

	_, err := repo.Availability(9999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
