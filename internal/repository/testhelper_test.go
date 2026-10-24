package repository

import (
	"os"
	"testing"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

// setupTestDB initializes a test database, runs migrations, and truncates tables.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	// Run migrations
	if err := RunMigrations(url, "../../migrations"); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	// Connect to DB
	db, err := NewPostgresDB(url)
	if err != nil {
		t.Fatalf("db: %v", err)
	}

	// Truncate tables
	if err := db.Exec("TRUNCATE TABLE loans, book_authors, copies, books, authors, categories, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}

	return db
}

// mustCategory creates a category with the given name and fails the test on error.
func mustCategory(t *testing.T, db *gorm.DB, name string) model.Category {
	t.Helper()
	category := model.Category{Name: name}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("failed to create category %s: %v", name, err)
	}
	return category
}

// mustAuthor creates an author with the given FullName and fails the test on error.
func mustAuthor(t *testing.T, db *gorm.DB, name string) model.Author {
	t.Helper()
	author := model.Author{
		FullName: name,
		Bio:      "",
	}
	if err := db.Create(&author).Error; err != nil {
		t.Fatalf("failed to create author %s: %v", name, err)
	}
	return author
}

// mustBook creates a book and associates it with authors, failing the test on error.
func mustBook(t *testing.T, db *gorm.DB, title, isbn string, categoryID uint, authorIDs []uint) model.Book {
	t.Helper()
	book := model.Book{
		Title:         title,
		ISBN:          isbn,
		CategoryID:    categoryID,
		Publisher:     "Test Publisher",
		PublishedYear: 2023,
		Description:   "Test description",
	}

	repo := NewBookRepository(db)
	if err := repo.Create(&book, authorIDs); err != nil {
		t.Fatalf("failed to create book %s: %v", title, err)
	}
	return book
}
