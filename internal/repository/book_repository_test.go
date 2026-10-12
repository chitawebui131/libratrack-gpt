package repository

import (
	"errors"
	"testing"
	"time"
	"sort"

	"libratrack/internal/model"
)

// Mock implementation for testing purposes
type InMemoryBookRepository struct {
	books map[uint]*model.Book
	nextID uint
}

func NewInMemoryBookRepository() *InMemoryBookRepository {
	return &InMemoryBookRepository{
		books: make(map[uint]*model.Book),
		nextID: 1,
	}
}

func (r *InMemoryBookRepository) Create(book *model.Book) error {
	book.ID = r.nextID
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()
	r.books[r.nextID] = book
	r.nextID++
	return nil
}

func (r *InMemoryBookRepository) FindByID(id uint) (*model.Book, error) {
	book, ok := r.books[id]
	if !ok {
		return nil, ErrNotFound
	}
	return book, nil
}

func (r *InMemoryBookRepository) List(category string, page, limit int) ([]*model.Book, error) {
	var results []*model.Book
	for _, book := range r.books {
		if category == "" || book.Category == category {
			results = append(results, book)
		}
	}

	// Sort by ID to ensure predictable pagination results
	sortedBooks := make([]*model.Book, len(results))
	copy(sortedBooks, results)
	sort.Slice(sortedBooks, func(i, j int) bool {
		return sortedBooks[i].ID < sortedBooks[j].ID
	})

	start := (page - 1) * limit
	end := start + limit

	if start >= len(sortedBooks) {
		return []*model.Book{}, nil
	}
	if end > len(sortedBooks) {
		end = len(sortedBooks)
	}

	return sortedBooks[start:end], nil
}

func (r *InMemoryBookRepository) Update(book *model.Book) error {
	if _, ok := r.books[book.ID]; !ok {
		return ErrNotFound
	}
	book.UpdatedAt = time.Now()
	r.books[book.ID] = book
	return nil
}

func (r *InMemoryBookRepository) Delete(id uint) error {
	if _, ok := r.books[id]; !ok {
		return ErrNotFound
	}
	delete(r.books, id)
	return nil
}

// newTestBook creates a book with the given title and category.
func newTestBook(title string, category string) *model.Book {
	return &model.Book{
		Title:    title,
		ISBN:     "isbn-" + title,
		Author:   "Author",
		Category: category,
	}
}

// TestCreateAndFindByID verifies that a book can be created and retrieved.
func TestCreateAndFindByID(t *testing.T) {
	repo := NewInMemoryBookRepository()
	book := newTestBook("Test Book", "Fiction")

	if err := repo.Create(book); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if book.ID != 1 {
		t.Fatalf("expected ID 1, got %d", book.ID)
	}

	if book.CreatedAt.IsZero() {
		t.Fatalf("CreatedAt is zero")
	}

	found, err := repo.FindByID(1)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}

	if found.Title != book.Title {
		t.Fatalf("expected title %s, got %s", book.Title, found.Title)
	}
}

// TestFindByIDNotFound verifies that FindByID returns ErrNotFound for a missing book.
func TestFindByIDNotFound(t *testing.T) {
	repo := NewInMemoryBookRepository()

	_, err := repo.FindByID(999)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestUpdate verifies that updating a book changes its title and updates UpdatedAt.
func TestUpdate(t *testing.T) {
	repo := NewInMemoryBookRepository()
	book := newTestBook("Original Title", "Fiction")
	if err := repo.Create(book); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	originalCreatedAt := book.CreatedAt

	book.Title = "Updated Title"
	if err := repo.Update(book); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	updated, err := repo.FindByID(book.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Fatalf("expected title %s, got %s", "Updated Title", updated.Title)
	}

	if updated.UpdatedAt.Before(originalCreatedAt) {
		t.Fatalf("UpdatedAt is before CreatedAt")
	}
}

// TestUpdateNotFound verifies that updating a non-existent book returns ErrNotFound.
func TestUpdateNotFound(t *testing.T) {
	repo := NewInMemoryBookRepository()
	book := &model.Book{
		ID:       999,
		Title:    "Nonexistent",
		ISBN:     "isbn-non",
		Author:   "Author",
		Category: "Fiction",
	}
	if err := repo.Update(book); err == nil {
		t.Fatalf("expected error, got nil")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestDelete verifies that deleting a book removes it from the repository.
func TestDelete(t *testing.T) {
	repo := NewInMemoryBookRepository()
	book := newTestBook("To Delete", "Fiction")
	if err := repo.Create(book); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := repo.Delete(book.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	_, err := repo.FindByID(book.ID)
	if err == nil {
		t.Fatalf("expected error after delete, got nil")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

// TestDeleteNotFound verifies that deleting a non-existent book returns ErrNotFound.
func TestDeleteNotFound(t *testing.T) {
	repo := NewInMemoryBookRepository()
	if err := repo.Delete(999); err == nil {
		t.Fatalf("expected error, got nil")
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestListCategoryFilter verifies that List filters by category correctly.
func TestListCategoryFilter(t *testing.T) {
	repo := NewInMemoryBookRepository()

	// Create two books in "scifi" category
	book1 := newTestBook("SciFi Book 1", "scifi")
	book2 := newTestBook("SciFi Book 2", "scifi")
	// Create one book in "history" category
	book3 := newTestBook("History Book", "history")

	if err := repo.Create(book1); err != nil {
		t.Fatalf("Create book1 returned error: %v", err)
	}
	if err := repo.Create(book2); err != nil {
		t.Fatalf("Create book2 returned error: %v", err)
	}
	if err := repo.Create(book3); err != nil {
		t.Fatalf("Create book3 returned error: %v", err)
	}

	// List by category "scifi"
	scifiBooks, err := repo.List("scifi", 1, 10)
	if err != nil {
		t.Fatalf("List scifi returned error: %v", err)
	}
	if len(scifiBooks) != 2 {
		t.Fatalf("expected 2 scifi books, got %d", len(scifiBooks))
	}

	// List all books
	allBooks, err := repo.List("", 1, 10)
	if err != nil {
		t.Fatalf("List all returned error: %v", err)
	}
	if len(allBooks) != 3 {
		t.Fatalf("expected 3 books, got %d", len(allBooks))
	}
}

// TestListPagination verifies that List paginates correctly.
func TestListPagination(t *testing.T) {
	repo := NewInMemoryBookRepository()

	// Create five books
	for i := 1; i <= 5; i++ {
		book := newTestBook("Book "+string(rune('0'+i)), "category")
		if err := repo.Create(book); err != nil {
			t.Fatalf("Create book %d returned error: %v", i, err)
		}
	}

	// Page 1, limit 2 -> IDs 1 and 2
	page1, err := repo.List("", 1, 2)
	if err != nil {
		t.Fatalf("List page 1 returned error: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 books on page 1, got %d", len(page1))
	}
	if page1[0].ID != 1 || page1[1].ID != 2 {
		t.Fatalf("expected IDs 1 and 2 on page 1, got %d and %d", page1[0].ID, page1[1].ID)
	}

	// Page 3, limit 2 -> ID 5
	page3, err := repo.List("", 3, 2)
	if err != nil {
		t.Fatalf("List page 3 returned error: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("expected 1 book on page 3, got %d", len(page3))
	}
	if page3[0].ID != 5 {
		t.Fatalf("expected ID 5 on page 3, got %d", page3[0].ID)
	}

	// Page 4, limit 2 -> empty slice
	page4, err := repo.List("", 4, 2)
	if err != nil {
		t.Fatalf("List page 4 returned error: %v", err)
	}
	if len(page4) != 0 {
		t.Fatalf("expected 0 books on page 4, got %d", len(page4))
	}
}
