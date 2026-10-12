package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func TestCopyCreateAndList(t *testing.T) {
	db := setupTestDB(t)

	// Create supporting data: category, author, book
	cat := &model.Category{Name: "Fiction"}
	if err := NewCategoryRepository(db).Create(cat); err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	author := &model.Author{Name: "John Doe"}
	if err := NewAuthorRepository(db).Create(author); err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	book := &model.Book{
		Title:      "Sample Book",
		CategoryID: cat.ID,
		AuthorID:   author.ID,
	}
	if err := NewBookRepository(db).Create(book); err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	copyRepo := NewCopyRepository(db)

	// First copy with empty status should default to "available"
	copy1 := &model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV-001",
		Status:      "",
	}
	if err := copyRepo.Create(copy1); err != nil {
		t.Fatalf("failed to create copy1: %v", err)
	}
	if copy1.Status != "available" {
		t.Fatalf("expected status 'available', got '%s'", copy1.Status)
	}

	// Second copy with status "lost"
	copy2 := &model.Copy{
		BookID:      book.ID,
		InventoryNo: "INV-002",
		Status:      "lost",
	}
	if err := copyRepo.Create(copy2); err != nil {
		t.Fatalf("failed to create copy2: %v", err)
	}

	// List copies by book
	copies, err := copyRepo.ListByBook(book.ID)
	if err != nil {
		t.Fatalf("failed to list copies: %v", err)
	}
	if len(copies) != 2 {
		t.Fatalf("expected 2 copies, got %d", len(copies))
	}
}

func TestCopyCreateUnknownBook(t *testing.T) {
	db := setupTestDB(t)
	copyRepo := NewCopyRepository(db)

	copy := &model.Copy{
		BookID:      9999,
		InventoryNo: "INV-999",
		Status:      "available",
	}
	err := copyRepo.Create(copy)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCopyDuplicateInventoryNo(t *testing.T) {
	db := setupTestDB(t)

	// Create supporting data
	cat := &model.Category{Name: "Non-Fiction"}
	if err := NewCategoryRepository(db).Create(cat); err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	author := &model.Author{Name: "Jane Smith"}
	if err := NewAuthorRepository(db).Create(author); err != nil {
		t.Fatalf("failed to create author: %v", err)
	}
	book := &model.Book{
		Title:      "Another Book",
		CategoryID: cat.ID,
		AuthorID:   author.ID,
	}
	if err := NewBookRepository(db).Create(book); err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	copyRepo := NewCopyRepository(db)

	copy1 := &model.Copy{
		BookID:      book.ID,
		InventoryNo: "DUP-001",
		Status:      "available",
	}
	if err := copyRepo.Create(copy1); err != nil {
		t.Fatalf("failed to create first copy: %v", err)
	}

	copy2 := &model.Copy{
		BookID:      book.ID,
		InventoryNo: "DUP-001",
		Status:      "available",
	}
	err := copyRepo.Create(copy2)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestAvailability(t *testing.T) {
	db := setupTestDB(t)

	// Create supporting data
	cat := &model.Category{Name: "Sci-Fi"}
	if err := NewCategoryRepository(db).Create(cat); err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	author := &model.Author{Name: "Alice"}
	if err := NewAuthorRepository(db).Create(author); err != nil {
		t.Fatalf("failed to create author: %v", err)
	}
	book := &model.Book{
		Title:      "Space Odyssey",
		CategoryID: cat.ID,
		AuthorID:   author.ID,
	}
	if err := NewBookRepository(db).Create(book); err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	copyRepo := NewCopyRepository(db)

	// 2 available
	for i := 1; i <= 2; i++ {
		c := &model.Copy{
			BookID:      book.ID,
			InventoryNo: "AVAIL-" + string(rune(i)),
			Status:      "available",
		}
		if err := copyRepo.Create(c); err != nil {
			t.Fatalf("failed to create available copy %d: %v", i, err)
		}
	}
	// 1 borrowed
	if err := copyRepo.Create(&model.Copy{
		BookID:      book.ID,
		InventoryNo: "BORROWED-1",
		Status:      "borrowed",
	}); err != nil {
		t.Fatalf("failed to create borrowed copy: %v", err)
	}
	// 1 lost
	if err := copyRepo.Create(&model.Copy{
		BookID:      book.ID,
		InventoryNo: "LOST-1",
		Status:      "lost",
	}); err != nil {
		t.Fatalf("failed to create lost copy: %v", err)
	}

	total, available, err := copyRepo.Availability(book.ID)
	if err != nil {
		t.Fatalf("availability returned error: %v", err)
	}
	if total != 4 {
		t.Fatalf("expected total 4, got %d", total)
	}
	if available != 2 {
		t.Fatalf("expected available 2, got %d", available)
	}
}

func TestAvailabilityNoCopies(t *testing.T) {
	db := setupTestDB(t)

	// Create supporting data
	cat := &model.Category{Name: "History"}
	if err := NewCategoryRepository(db).Create(cat); err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	author := &model.Author{Name: "Bob"}
	if err := NewAuthorRepository(db).Create(author); err != nil {
		t.Fatalf("failed to create author: %v", err)
	}
	book := &model.Book{
		Title:      "Ancient Times",
		CategoryID: cat.ID,
		AuthorID:   author.ID,
	}
	if err := NewBookRepository(db).Create(book); err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	copyRepo := NewCopyRepository(db)

	total, available, err := copyRepo.Availability(book.ID)
	if err != nil {
		t.Fatalf("availability returned error: %v", err)
	}
	if total != 0 || available != 0 {
		t.Fatalf("expected 0,0 for no copies, got %d,%d", total, available)
	}
}

func TestAvailabilityUnknownBook(t *testing.T) {
	db := setupTestDB(t)
	copyRepo := NewCopyRepository(db)

	_, _, err := copyRepo.Availability(9999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
