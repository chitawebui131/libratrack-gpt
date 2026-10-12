package repository

import "testing"

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
	mustBook(t, db, "Book 1", "ISBN-1", cat.ID