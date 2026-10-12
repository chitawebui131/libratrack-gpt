package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestBookCreateAndFind(t *testing.T) {
	db := setupTestDB(t)

	// 1. Setup Category
	category := mustCategory(t, db, "Fiction")

	// 2. Setup Authors
	author1 := mustAuthor(t, db, "Author One")
	author2 := mustAuthor(t, db, "Author Two")

	// 3. Create Book
	book := &model.Book{
		Title:         "Test Book Title",
		ISBN:          "1234567890",
		CategoryID:    category.ID,
		Publisher:     "Test Pub",
		PublishedYear: 2020,
		Description:   "A test book",
	}
	authorIDs := []uint{author1.ID, author2.ID}

	repo := NewBookRepository(db)
	err := repo.Create(book, authorIDs)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 4. FindByID and verify
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

	if found.Category == nil || found.Category.Name