package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestBookUpdate(t *testing.T) {
	db := setupTestDB(t)

	// Create category and authors
	category := mustCategory(t, db, "Fiction")
	author1 := mustAuthor(t, db, "Author One")
	author2 := mustAuthor(t, db, "Author Two")

	// Create a book with the first author
	book := mustBook(t, db, "Original Title", "ISBN-1234", category.ID, []uint{author1.ID})

	// Update the book: new title, same ISBN, same category, second author only
	repo := NewBookRepository(db)
	if err := repo.Update(&model.Book{
		ID:         book.ID,
		Title:      "New Title",
		ISBN:       book.ISBN,
		CategoryID: category.ID,
	}, []uint{author2.ID}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	// Verify updated book
	updated, err := repo.FindByID(book.ID)
	if err != nil {
		t.Fatalf("FindByID after update returned error: %v", err)
	}
	if updated.Title != "New Title" {
		t.Fatalf("Expected title 'New Title', got '%s'", updated.Title)
	}
	if len(updated.Authors) != 1 {
		t.Fatalf("Expected 1 author after update, got %d", len(updated.Authors))
	}
	if updated.Authors[0].ID != author2.ID {
		t.Fatalf("Expected author ID %d, got %d", author2.ID, updated.Authors[0].ID)
	}
}

func TestBookUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)

	// Create a category to satisfy foreign key
	category := mustCategory(t, db, "Fiction")

	repo := NewBookRepository(db)
	err := repo.Update(&model.Book{
		ID:         9999,
		Title:      "Nonexistent",
		ISBN:       "ISBN-9999",
		CategoryID: category.ID,
	}, []uint{})
	if err == nil {
		t.Fatalf("Expected error for non-existent book, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestBookDelete(t *testing.T) {
	db := setupTestDB(t)

	// Create category and author
	category := mustCategory(t, db, "Fiction")
	author := mustAuthor(t, db, "Author One")

	// Create a book
	book := mustBook(t, db, "To Delete", "ISBN-DEL", category.ID, []uint{author.ID})

	// Insert a copy for the book
	if err := db.Create(&model.Copy{BookID: book.ID, InventoryNo: "INV-T1"}).Error; err != nil {
		t.Fatalf("Failed to create copy: %v", err)
	}

	repo := NewBookRepository(db)
	if err := repo.Delete(book.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	// Verify book is deleted
	if _, err := repo.FindByID(book.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expected ErrNotFound after delete, got %v", err)
	}

	// Verify copies are cascaded
	var n int64
	if err := db.Model(&model.Copy{}).Where("book_id = ?", book.ID).Count(&n).Error; err != nil {
		t.Fatalf("Failed to count copies: %v", err)
	}
	if n != 0 {
		t.Fatalf("Expected 0 copies after book deletion, got %d", n)
	}
}

func TestBookDeleteNotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewBookRepository(db)
	err := repo.Delete(9999)
	if err == nil {
		t.Fatalf("Expected error for deleting non-existent book, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}
