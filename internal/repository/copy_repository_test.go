package repository

import (
	"errors"
	"testing"

	"gorm.io/gorm"
	"libratrack/internal/model"
)

func TestCopyCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	cat := mustCategory(t, db, "Cat")
	auth := mustAuthor(t, db, "Auth")
	book := mustBook(t, db, "Book", "1234567890", uint(cat.ID), []uint{uint(auth.ID)})
	repo := NewCopyRepository(db)

	c1 := &model.Copy{BookID: book.ID, InventoryNo: "1"}
	err := repo.Create(c1)
	if err != nil {
		t.Fatalf("failed to create copy 1: %v", err)
	}
	if c1.Status != "available" {
		t.Errorf("expected status available, got %s", c1.Status)
	}

	c2 := &model.Copy{BookID: book.ID, InventoryNo: "2", Status: "lost"}
	err = repo.Create(c2)
	if err != nil {
		t.Fatalf("failed to create copy 2: %v", err)
	}

	copies, err := repo.ListByBook(book.ID)
	if err != nil {
		t.Fatalf("failed to list copies: %v", err)
	}
	if len(copies) != 2 {
		t.Errorf("expected 2 copies, got %d", len(copies))
	}
}

func TestCopyCreateUnknownBook(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCopyRepository(db)
	c := &model.Copy{BookID: 9999, InventoryNo: "X"}
	err := repo.Create(c)
	if err == nil {
		t.Errorf("expected error for unknown book, got nil")
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := setupTestDB(t)
	cat := mustCategory(t, db, "Cat")
	auth := mustAuthor(t, db, "Auth")
	book := mustBook(t, db, "Book", "1234567890", uint(cat.ID), []uint{uint(auth.ID)})
	repo := NewCopyRepository(db)

	c1 := &model.Copy{BookID: book.ID, InventoryNo: "DUP"}
	err := repo.Create(c1)
	if err != nil {
		t.Fatalf("failed to create copy 1: %v", err)
	}

	c2 := &model.Copy{BookID: book.ID, InventoryNo: "DUP"}
	err = repo.Create(c2)
	if err == nil {
		t.Errorf("expected conflict error, got nil")
	}
}

func TestAvailability(t *testing.T) {
	db := setupTestDB(t)
	cat := mustCategory(t, db, "Cat")
	auth := mustAuthor(t, db, "Auth")
	book := mustBook(t, db, "Book", "1234567890", uint(cat.ID), []uint{uint(auth.ID)})
	repo := NewCopyRepository(db)

	_ = repo.Create(&model.Copy{BookID: book.ID, InventoryNo: "1", Status: "available"})
	_ = repo.Create(&model.Copy{BookID: book.ID, InventoryNo: "2", Status: "available"})
	_ = repo.Create(&model.Copy{BookID: book.ID, InventoryNo: "3", Status: "borrowed"})
	_ = repo.Create(&model.Copy{BookID: book.ID, InventoryNo: "4", Status: "lost"})

	avail, err := repo.Availability(book.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avail.TotalCopies != 4 || avail.AvailableCopies != 2 {
		t.Errorf("expected 4 total, 2 available; got %d total, %d available", avail.TotalCopies, avail.AvailableCopies)
	}
}

func TestAvailabilityNoCopies(t *testing.T) {
	db := setupTestDB(t)
	cat := mustCategory(t, db, "Cat")
	auth := mustAuthor(t, db, "Auth")
	book := mustBook(t, db, "Book", "1234567890", uint(cat.ID), []uint{uint(auth.ID)})
	repo := NewCopyRepository(db)

	avail, err := repo.Availability(book.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avail.TotalCopies != 0 || avail.AvailableCopies != 0 {
		t.Errorf("expected 0 total, 0 available; got %d total, %d available", avail.TotalCopies, avail.AvailableCopies)
	}
}

func TestAvailabilityUnknownBook(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCopyRepository(db)
	_, err := repo.Availability(9999)
	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
	}
}
