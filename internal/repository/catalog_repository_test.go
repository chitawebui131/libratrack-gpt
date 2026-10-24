package repository

import (
	"testing"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

// seedCatalog creates one category, two authors and three books:
//
//	"Go Programming" (ISBN 9780000000001, category, author "Alan Donovan", one available copy),
//	"Sailing 100% Fun" (ISBN 9780000000002, category, author "Jane Sailor", one borrowed copy),
//	"Cooking Basics" (ISBN 9780000000003, category, author "Jane Sailor", NO copies).
//
// It returns the repository and the ids of the three books.
func seedCatalog(t *testing.T, db *gorm.DB) (CatalogRepository, []uint) {
	t.Helper()
	category := model.Category{Name: "Test Category"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	author1 := model.Author{FullName: "Alan Donovan"}
	if err := db.Create(&author1).Error; err != nil {
		t.Fatalf("failed to create author Alan Donovan: %v", err)
	}

	author2 := model.Author{FullName: "Jane Sailor"}
	if err := db.Create(&author2).Error; err != nil {
		t.Fatalf("failed to create author Jane Sailor: %v", err)
	}

	book1 := model.Book{
		Title:         "Go Programming",
		ISBN:          "9780000000001",
		CategoryID:    category.ID,
		Publisher:     "Test Publisher",
		PublishedYear: 2023,
		Description:   "Test description",
		Authors:       []model.Author{author1},
	}
	repo := NewBookRepository(db)
	if err := repo.Create(&book1, []uint{author1.ID}); err != nil {
		t.Fatalf("failed to create book Go Programming: %v", err)
	}

	copy1 := model.Copy{
		BookID:      book1.ID,
		Status:      model.CopyStatusAvailable,
		InventoryNo: "INV001",
	}
	if err := db.Create(&copy1).Error; err != nil {
		t.Fatalf("failed to create copy for Go Programming: %v", err)
	}

	book2 := model.Book{
		Title:         "Sailing 100% Fun",
		ISBN:          "9780000000002",
		CategoryID:    category.ID,
		Publisher:     "Test Publisher",
		PublishedYear: 2023,
		Description:   "Test description",
		Authors:       []model.Author{author2},
	}
	if err := repo.Create(&book2, []uint{author2.ID}); err != nil {
		t.Fatalf("failed to create book Sailing 100 Fun: %v", err)
	}

	copy2 := model.Copy{
		BookID:      book2.ID,
		Status:      "borrowed",
		InventoryNo: "INV002",
	}
	if err := db.Create(&copy2).Error; err != nil {
		t.Fatalf("failed to create copy for Sailing 100 Fun: %v", err)
	}

	book3 := model.Book{
		Title:         "Cooking Basics",
		ISBN:          "9780000000003",
		CategoryID:    category.ID,
		Publisher:     "Test Publisher",
		PublishedYear: 2023,
		Description:   "Test description",
		Authors:       []model.Author{author2},
	}
	if err := repo.Create(&book3, []uint{author2.ID}); err != nil {
		t.Fatalf("failed to create book Cooking Basics: %v", err)
	}

	return NewCatalogRepository(db), []uint{book1.ID, book2.ID, book3.ID}
}

func TestSearchByTitleISBNAndAuthor(t *testing.T) {
	db := setupTestDB(t)
	repo, bookIDs := seedCatalog(t, db)

	// Query "go prog" -> 1 book
	filter := BookFilter{Query: "go prog"}
	books, total, err := repo.Search(filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("Expected 1 book, got %d", len(books))
	}
	if books[0].ID != bookIDs[0] {
		t.Fatalf("Expected book ID %d, got %d", bookIDs[0], books[0].ID)
	}
	if total != 1 {
		t.Fatalf("Expected total 1, got %d", total)
	}

	// Query "9780000000002" -> 1 book
	filter = BookFilter{Query: "9780000000002"}
	books, total, err = repo.Search(filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("Expected 1 book, got %d", len(books))
	}
	if books[0].ID != bookIDs[1] {
		t.Fatalf("Expected book ID %d, got %d", bookIDs[1], books[0].ID)
	}
	if total != 1 {
		t.Fatalf("Expected total 1, got %d", total)
	}

	// Query "alan" -> 1 book (author match)
	filter = BookFilter{Query: "alan"}
	books, total, err = repo.Search(filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("Expected 1 book, got %d", len(books))
	}
	if books[0].ID != bookIDs[0] {
		t.Fatalf("Expected book ID %d, got %d", bookIDs[0], books[0].ID)
	}
	if total != 1 {
		t.Fatalf("Expected total 1, got %d", total)
	}
}

func TestSearchWildcardIsEscaped(t *testing.T) {
	db := setupTestDB(t)
	repo, _ := seedCatalog(t, db)

	// Query "%" must return only the book whose title contains a literal percent sign ("Sailing 100% Fun"), NOT all books.
	filter := BookFilter{Query: "%"}
	books, total, err := repo.Search(filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("Expected 1 book, got %d", len(books))
	}
	if books[0].Title != "Sailing 100% Fun" {
		t.Fatalf("Expected book title 'Sailing 100%% Fun', got '%s'", books[0].Title)
	}
	if total != 1 {
		t.Fatalf("Expected total 1, got %d", total)
	}
}

func TestSearchFiltersAndPaging(t *testing.T) {
	db := setupTestDB(t)
	repo, bookIDs := seedCatalog(t, db)

	// OnlyAvailable true -> only "Go Programming"
	filter := BookFilter{OnlyAvailable: true}
	books, total, err := repo.Search(filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("Expected 1 book, got %d", len(books))
	}
	if books[0].ID != bookIDs[0] {
		t.Fatalf("Expected book ID %d, got %d", bookIDs[0], books[0].ID)
	}
	if total != 1 {
		t.Fatalf("Expected total 1, got %d", total)
	}

	// AuthorID of "Jane Sailor" -> 2 books
	author := model.Author{FullName: "Jane Sailor"}
	if err := db.Where("full_name = ?", author.FullName).First(&author).Error; err != nil {
		t.Fatalf("failed to find author Jane Sailor: %v", err)
	}
	filter = BookFilter{AuthorID: author.ID}
	books, total, err = repo.Search(filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(books) != 2 {
		t.Fatalf("Expected 2 books, got %d", len(books))
	}
	if total != 2 {
		t.Fatalf("Expected total 2, got %d", total)
	}

	// Page 1 Limit 1 over all books -> 1 book and total 3
	filter = BookFilter{Page: 1, Limit: 1}
	books, total, err = repo.Search(filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("Expected 1 book, got %d", len(books))
	}
	if total != 3 {
		t.Fatalf("Expected total 3, got %d", total)
	}

	// Page 2 Limit 1 -> a different book
	filter = BookFilter{Page: 2, Limit: 1}
	books, total, err = repo.Search(filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("Expected 1 book, got %d", len(books))
	}
	if total != 3 {
		t.Fatalf("Expected total 3, got %d", total)
	}
}

func TestAvailabilityForBooks(t *testing.T) {
	db := setupTestDB(t)
	repo, bookIDs := seedCatalog(t, db)

	// For the three book ids: Go Programming Total 1 Available 1; Sailing Total 1 Available 0; Cooking Total 0 Available 0 (present in the map)
	availabilities, err := repo.AvailabilityForBooks(bookIDs)
	if err != nil {
		t.Fatalf("AvailabilityForBooks failed: %v", err)
	}
	if len(availabilities) != 3 {
		t.Fatalf("Expected 3 availabilities, got %d", len(availabilities))
	}

	// Check the availability values - need to access through the correct struct fields
	goProgAvail := availabilities[bookIDs[0]]
	if goProgAvail.Total != 1 || goProgAvail.Available != 1 {
		t.Fatalf("Expected Go Programming availability Total 1 Available 1, got Total %d Available %d", goProgAvail.Total, goProgAvail.Available)
	}

	sailingAvail := availabilities[bookIDs[1]]
	if sailingAvail.Total != 1 || sailingAvail.Available != 0 {
		t.Fatalf("Expected Sailing availability Total 1 Available 0, got Total %d Available %d", sailingAvail.Total, sailingAvail.Available)
	}

	cookingAvail := availabilities[bookIDs[2]]
	if cookingAvail.Total != 0 || cookingAvail.Available != 0 {
		t.Fatalf("Expected Cooking availability Total 0 Available 0, got Total %d Available %d", cookingAvail.Total, cookingAvail.Available)
	}

	// AvailabilityForBooks(nil) returns an empty map
	availabilities, err = repo.AvailabilityForBooks(nil)
	if err != nil {
		t.Fatalf("AvailabilityForBooks failed: %v", err)
	}
	if len(availabilities) != 0 {
		t.Fatalf("Expected 0 availabilities, got %d", len(availabilities))
	}
}
