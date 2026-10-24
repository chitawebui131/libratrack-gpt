package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestBookCreateAndFind(t *testing.T) {
	db := setupTestDB(t)

	// Create category
	category := mustCategory(t, db, "Fiction")

	// Create authors
	author1 := mustAuthor(t, db, "Author One")
	author2 := mustAuthor(t, db, "Author Two")

	// Create book via repository
	repo := NewBookRepository(db)
	book := &model.Book{
		Title:         "Test Book Title",
		ISBN:          "1234567890",
		CategoryID:    category.ID,
		Publisher:     "Test Pub",
		PublishedYear: 2020,
		Description:   "A test book",
	}
	authorIDs := []uint{author1.ID, author2.ID}
	if err := repo.Create(book, authorIDs); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Find by ID
	found, err := repo.FindByID(book.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.Title != book.Title {
		t.Errorf("Expected title %s, got %s", book.Title, found.Title)
	}

	if len(found.Authors) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(found.Authors))
	}

	if found.Category == nil || found.Category.Name != category.Name {
		t.Errorf("Expected category %s, got %+v", category.Name, found.Category)
	}
}

func TestBookCreateDuplicateISBN(t *testing.T) {
	db := setupTestDB(t)

	// Setup category and authors
	category := mustCategory(t, db, "Fiction")
	author := mustAuthor(t, db, "Author One")

	// First book
	repo := NewBookRepository(db)
	book1 := &model.Book{
		Title:         "First Book",
		ISBN:          "DUPISBN",
		CategoryID:    category.ID,
		Publisher:     "Pub",
		PublishedYear: 2021,
		Description:   "First",
	}
	if err := repo.Create(book1, []uint{author.ID}); err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	// Second book with same ISBN
	book2 := &model.Book{
		Title:         "Second Book",
		ISBN:          "DUPISBN",
		CategoryID:    category.ID,
		Publisher:     "Pub",
		PublishedYear: 2022,
		Description:   "Second",
	}
	err := repo.Create(book2, []uint{author.ID})
	if err == nil {
		t.Fatalf("Expected error for duplicate ISBN, got nil")
	}
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("Expected ErrConflict, got %v", err)
	}
}

func TestBookCreateInvalidCategory(t *testing.T) {
	db := setupTestDB(t)

	// Setup author
	author := mustAuthor(t, db, "Author One")

	repo := NewBookRepository(db)
	book := &model.Book{
		Title:         "Book with bad category",
		ISBN:          "ISBN123",
		CategoryID:    9999, // non-existent
		Publisher:     "Pub",
		PublishedYear: 2023,
		Description:   "Bad category",
	}
	err := repo.Create(book, []uint{author.ID})
	if err == nil {
		t.Fatalf("Expected error for invalid category, got nil")
	}
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("Expected ErrInvalidReference, got %v", err)
	}
}

func TestBookCreateInvalidAuthor(t *testing.T) {
	db := setupTestDB(t)

	// Setup category
	category := mustCategory(t, db, "Fiction")

	repo := NewBookRepository(db)
	book := &model.Book{
		Title:         "Book with bad author",
		ISBN:          "ISBN456",
		CategoryID:    category.ID,
		Publisher:     "Pub",
		PublishedYear: 2024,
		Description:   "Bad author",
	}
	invalidAuthorID := uint(9999)
	err := repo.Create(book, []uint{invalidAuthorID})
	if err == nil {
		t.Fatalf("Expected error for invalid author, got nil")
	}
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("Expected ErrInvalidReference, got %v", err)
	}

	// Verify no author with id 9999 exists
	var n int64
	db.Model(&model.Author{}).Where("id = ?", invalidAuthorID).Count(&n)
	if n != 0 {
		t.Fatalf("Expected no author with id %d, found %d", invalidAuthorID, n)
	}
}

func TestBookFindNotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewBookRepository(db)
	_, err := repo.FindByID(9999)
	if err == nil {
		t.Fatalf("Expected error for not found, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}
