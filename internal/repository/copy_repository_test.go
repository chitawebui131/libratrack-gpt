package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
	"libratrack/internal/repository/testhelper"
)

func TestCopyCreateAndList(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(db)

	// Create category and author for the book
	category := model.Category{
		Name: "Fiction",
	}
	testhelper.CreateCategory(t, db, &category)

	author := model.Author{
		FirstName: "John",
		LastName:  "Doe",
	}
	testhelper.CreateAuthor(t, db, &author)

	// Create a book
	book := model.Book{
		Title:       "Test Book",
		Description: "A test book",
		CategoryID:  category.ID,
		Authors:     []model.Author{author},
	}
	testhelper.CreateBook(t, db, &book)

	// Create two copies
	copy1 := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV-001",
		Status:      "", // Empty status should default to "available"
	}
	copyRepo := NewCopyRepository(db)
	createdCopy1, err := copyRepo.Create(copy1)
	if err != nil {
		t.Fatalf("Failed to create first copy: %v", err)
	}

	// Check that the first copy got the status "available" automatically
	if createdCopy1.Status != model.CopyStatusAvailable {
		t.Errorf("Expected first copy status to be %s, got %s", model.CopyStatusAvailable, createdCopy1.Status)
	}

	copy2 := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV-002",
		Status:      model.CopyStatusLost,
	}
	createdCopy2, err := copyRepo.Create(copy2)
	if err != nil {
		t.Fatalf("Failed to create second copy: %v", err)
	}

	// Check that the second copy has the correct status
	if createdCopy2.Status != model.CopyStatusLost {
		t.Errorf("Expected second copy status to be %s, got %s", model.CopyStatusLost, createdCopy2.Status)
	}

	// List copies by book
	copies, err := copyRepo.ListByBook(book.ID)
	if err != nil {
		t.Fatalf("Failed to list copies: %v", err)
	}

	if len(copies) != 2 {
		t.Errorf("Expected 2 copies, got %d", len(copies))
	}
}

func TestCopyCreateUnknownBook(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(db)

	copyRepo := NewCopyRepository(db)
	copy := model.Copy{
		BookID:      9999, // Non-existent book ID
		InventoryNo: "INV-001",
	}
	_, err := copyRepo.Create(copy)
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(db)

	// Create category and author for the book
	category := model.Category{
		Name: "Fiction",
	}
	testhelper.CreateCategory(t, db, &category)

	author := model.Author{
		FirstName: "John",
		LastName:  "Doe",
	}
	testhelper.CreateAuthor(t, db, &author)

	// Create a book
	book := model.Book{
		Title:       "Test Book",
		Description: "A test book",
		CategoryID:  category.ID,
		Authors:     []model.Author{author},
	}
	testhelper.CreateBook(t, db, &book)

	// Create first copy
	copyRepo := NewCopyRepository(db)
	firstCopy := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV-001",
	}
	_, err := copyRepo.Create(firstCopy)
	if err != nil {
		t.Fatalf("Failed to create first copy: %v", err)
	}

	// Try to create a second copy with the same inventory number
	secondCopy := model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV-001",
	}
	_, err = copyRepo.Create(secondCopy)
	if !errors.Is(err, model.ErrConflict) {
		t.Errorf("Expected ErrConflict for duplicate inventory number, got %v", err)
	}
}

func TestAvailability(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(db)

	// Create category and author for the book
	category := model.Category{
		Name: "Fiction",
	}
	testhelper.CreateCategory(t, db, &category)

	author := model.Author{
		FirstName: "John",
		LastName:  "Doe",
	}
	testhelper.CreateAuthor(t, db, &author)

	// Create a book
	book := model.Book{
		Title:       "Test Book",
		Description: "A test book",
		CategoryID:  category.ID,
		Authors:     []model.Author{author},
	}
	testhelper.CreateBook(t, db, &book)

	// Create 4 copies with different statuses
	copyRepo := NewCopyRepository(db)
	copies := []model.Copy{
		{BookID: book.ID, InventoryNo: "INV-001", Status: model.CopyStatusAvailable},
		{BookID: book.ID, InventoryNo: "INV-002", Status: model.CopyStatusAvailable},
		{BookID: book.ID, InventoryNo: "INV-003", Status: model.CopyStatusBorrowed},
		{BookID: book.ID, InventoryNo: "INV-004", Status: model.CopyStatusLost},
	}
	for _, c := range copies {
		_, err := copyRepo.Create(c)
		if err != nil {
			t.Fatalf("Failed to create copy: %v", err)
		}
	}

	// Test availability
	availability, err := copyRepo.Availability(book.ID)
	if err != nil {
		t.Fatalf("Failed to get availability: %v", err)
	}

	if availability.TotalCopies != 4 {
		t.Errorf("Expected TotalCopies to be 4, got %d", availability.TotalCopies)
	}

	if availability.AvailableCopies != 2 {
		t.Errorf("Expected AvailableCopies to be 2, got %d", availability.AvailableCopies)
	}
}

func TestAvailabilityNoCopies(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(db)

	// Create category and author for the book
	category := model.Category{
		Name: "Fiction",
	}
	testhelper.CreateCategory(t, db, &category)

	author := model.Author{
		FirstName: "John",
		LastName:  "Doe",
	}
	testhelper.CreateAuthor(t, db, &author)

	// Create a book
	book := model.Book{
		Title:       "Test Book",
		Description: "A test book",
		CategoryID:  category.ID,
		Authors:     []model.Author{author},
	}
	testhelper.CreateBook(t, db, &book)

	// Test availability for book with no copies
	availability, err := copyRepo.Availability(book.ID)
	if err != nil {
		t.Fatalf("Failed to get availability: %v", err)
	}

	if availability.TotalCopies != 0 {
		t.Errorf("Expected TotalCopies to be 0, got %d", availability.TotalCopies)
	}

	if availability.AvailableCopies != 0 {
		t.Errorf("Expected AvailableCopies to be 0, got %d", availability.AvailableCopies)
	}
}

func TestAvailabilityUnknownBook(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(db)

	copyRepo := NewCopyRepository(db)
	_, err := copyRepo.Availability(9999) // Non-existent book ID
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
