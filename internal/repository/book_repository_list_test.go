package repository

import (
	"testing"

	"libratrack/internal/model"
)

func TestBookListByCategory(t *testing.T) {
	db := setupTestDB(t)

	// Create categories
	catA := mustCategory(t, db, "A")
	catB := mustCategory(t, db, "B")

	// Create an author
	author := mustAuthor(t, db, "Author One")

	// Create books in category A
	mustBook(t, db, "Book A1", "ISBN-A1", catA.ID, []uint{author.ID})
	mustBook(t, db, "Book A2", "ISBN-A2", catA.ID, []uint{author.ID})
	mustBook(t, db, "Book A3", "ISBN-A3", catA.ID, []uint{author.ID})

	// Create books in category B
	mustBook(t, db, "Book B1", "ISBN-B1", catB.ID, []uint{author.ID})
	mustBook(t, db, "Book B2", "ISBN-B2", catB.ID, []uint{author.ID})

	repo := NewBookRepository(db)

	// List books in category A
	booksA, err := repo.List(catA.ID, 1, 20)
	if err != nil {
		t.Fatalf("List for category A returned error: %v", err)
	}
	if len(booksA) != 3 {
		t.Fatalf("expected 3 books in category A, got %d", len(booksA))
	}

	// List books in category B
	booksB, err := repo.List(catB.ID, 1, 20)
	if err != nil {
		t.Fatalf("List for category B returned error: %v", err)
	}
	if len(booksB) != 2 {
		t.Fatalf("expected 2 books in category B, got %d", len(booksB))
	}

	// List all books
	booksAll, err := repo.List(0, 1, 20)
	if err != nil {
		t.Fatalf("List for all books returned error: %v", err)
	}
	if len(booksAll) != 5 {
		t.Fatalf("expected 5 books in total, got %d", len(booksAll))
	}
}

func TestBookListPagination(t *testing.T) {
	db := setupTestDB(t)

	// Create a category and an author
	cat := mustCategory(t, db, "Cat")
	author := mustAuthor(t, db, "Author One")

	// Create five books in the same category
	mustBook(t, db, "Book 1", "ISBN-1", cat.ID, []uint{author.ID})
	mustBook(t, db, "Book 2", "ISBN-2", cat.ID, []uint{author.ID})
	mustBook(t, db, "Book 3", "ISBN-3", cat.ID, []uint{author.ID})
	mustBook(t, db, "Book 4", "ISBN-4", cat.ID, []uint{author.ID})
	mustBook(t, db, "Book 5", "ISBN-5", cat.ID, []uint{author.ID})

	repo := NewBookRepository(db)

	// Page 1, limit 2 -> IDs 1 and 2
	booksPage1, err := repo.List(0, 1, 2)
	if err != nil {
		t.Fatalf("List page 1 returned error: %v", err)
	}
	if len(booksPage1) != 2 {
		t.Fatalf("expected 2 books on page 1, got %d", len(booksPage1))
	}

	// Page 3, limit 2 -> ID 5
	booksPage3, err := repo.List(0, 3, 2)
	if err != nil {
		t.Fatalf("List page 3 returned error: %v", err)
	}
	if len(booksPage3) != 1 {
		t.Fatalf("expected 1 book on page 3, got %d", len(booksPage3))
	}

	// Page 4, limit 2 -> empty slice
	booksPage4, err := repo.List(0, 4, 2)
	if err != nil {
		t.Fatalf("List page 4 returned error: %v", err)
	}
	if len(booksPage4) != 0 {
		t.Fatalf("expected 0 books on page 4, got %d", len(booksPage4))
	}
}

func TestBookListLoadsRelations(t *testing.T) {
	db := setupTestDB(t)

	// Create categories
	cat1 := mustCategory(t, db, "Fiction")
	cat2 := mustCategory(t, db, "Science")

	// Create authors
	author1 := mustAuthor(t, db, "Author One")
	author2 := mustAuthor(t, db, "Author Two")

	// Create books with different author combinations
	book1 := mustBook(t, db, "Book 1", "ISBN-1", cat1.ID, []uint{author1.ID})
	book2 := mustBook(t, db, "Book 2", "ISBN-2", cat1.ID, []uint{author1.ID, author2.ID})
	book3 := mustBook(t, db, "Book 3", "ISBN-3", cat2.ID, []uint{author2.ID})

	repo := NewBookRepository(db)

	// List all books
	books, err := repo.List(0, 1, 20)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	// Verify each book has correct relations
	if len(books) != 3 {
		t.Fatalf("expected 3 books, got %d", len(books))
	}

	// Find the books by ID to check their relations
	var foundBook1, foundBook2, foundBook3 *model.Book
	for _, b := range books {
		switch b.ID {
		case book1.ID:
			foundBook1 = &b
		case book2.ID:
			foundBook2 = &b
		case book3.ID:
			foundBook3 = &b
		}
	}

	if foundBook1 == nil || foundBook2 == nil || foundBook3 == nil {
		t.Fatal("Could not find all books in the list")
	}

	// Check book1 relations
	if foundBook1.Category == nil {
		t.Error("Book1 should have a non-nil Category")
	}
	if foundBook1.Category.ID != cat1.ID {
		t.Errorf("Book1 category ID mismatch: expected %d, got %d", cat1.ID, foundBook1.Category.ID)
	}
	if len(foundBook1.Authors) != 1 {
		t.Errorf("Book1 should have exactly 1 author, got %d", len(foundBook1.Authors))
	}
	if foundBook1.Authors[0].ID != author1.ID {
		t.Errorf("Book1 author ID mismatch: expected %d, got %d", author1.ID, foundBook1.Authors[0].ID)
	}

	// Check book2 relations
	if foundBook2.Category == nil {
		t.Error("Book2 should have a non-nil Category")
	}
	if foundBook2.Category.ID != cat1.ID {
		t.Errorf("Book2 category ID mismatch: expected %d, got %d", cat1.ID, foundBook2.Category.ID)
	}
	if len(foundBook2.Authors) != 2 {
		t.Errorf("Book2 should have exactly 2 authors, got %d", len(foundBook2.Authors))
	}

	// Check book3 relations
	if foundBook3.Category == nil {
		t.Error("Book3 should have a non-nil Category")
	}
	if foundBook3.Category.ID != cat2.ID {
		t.Errorf("Book3 category ID mismatch: expected %d, got %d", cat2.ID, foundBook3.Category.ID)
	}
	if len(foundBook3.Authors) != 1 {
		t.Errorf("Book3 should have exactly 1 author, got %d", len(foundBook3.Authors))
	}
	if foundBook3.Authors[0].ID != author2.ID {
		t.Errorf("Book3 author ID mismatch: expected %d, got %d", author2.ID, foundBook3.Authors[0].ID)
	}
}
