package repository

import (
	"errors"
	"testing"

	"libratrack/internal/model"
)

func sampleBook() model.ImportBook {
	return model.ImportBook{LegacyCode: 35, Title: "Test Book", Authors: []string{"Zed Last", "Ann First"}, Year: 1987, City: "Kyiv", Publisher: "Pub", Department: "3", Pages: 232, BBK: "34.63", UDK: "821", ISBN: "9780306406157", Language: "ukr", CategoryName: "Test Category", Inventory: []string{"9001", "9002", "9003"}}
}

func TestUpsertBookCreates(t *testing.T) {
	db := setupTestDB(t)
	repo := NewImportRepository(db)
	res, err := repo.UpsertBook(sampleBook())
	if err != nil {
		t.Fatalf("UpsertBook failed: %v", err)
	}

	if !res.BookCreated {
		t.Error("Expected BookCreated to be true")
	}

	if res.CopiesAdded != 3 {
		t.Errorf("Expected CopiesAdded to be 3, got %d", res.CopiesAdded)
	}

	if res.CopiesSkipped != 0 {
		t.Errorf("Expected CopiesSkipped to be 0, got %d", res.CopiesSkipped)
	}

	// Check database state
	var book model.Book
	if err := db.Where("legacy_code = ?", 35).First(&book).Error; err != nil {
		t.Fatalf("Failed to find book: %v", err)
	}

	if book.ISBN != "9780306406157" {
		t.Errorf("Expected ISBN to be '9780306406157', got '%s'", book.ISBN)
	}

	if book.UDK != "821" {
		t.Errorf("Expected UDK to be '821', got '%s'", book.UDK)
	}

	if book.Language != "ukr" {
		t.Errorf("Expected Language to be 'ukr', got '%s'", book.Language)
	}

	var copies []model.Copy
	if err := db.Where("book_id = ? AND department = ?", book.ID, "3").Find(&copies).Error; err != nil {
		t.Fatalf("Failed to find copies: %v", err)
	}

	if len(copies) != 3 {
		t.Errorf("Expected 3 copies, got %d", len(copies))
	}

	for _, copy := range copies {
		if copy.Status != model.CopyStatusAvailable {
			t.Errorf("Expected copy status to be 'available', got '%s'", copy.Status)
		}
		if copy.Department != "3" {
			t.Errorf("Expected copy department to be '3', got '%s'", copy.Department)
		}
	}

	var authors []model.Author
	if err := db.Joins("JOIN book_authors ON authors.id = book_authors.author_id WHERE book_authors.book_id = ?", book.ID).Find(&authors).Error; err != nil {
		t.Fatalf("Failed to find authors: %v", err)
	}

	if len(authors) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(authors))
	}

	var category model.Category
	if err := db.Where("name = ?", "Test Category").First(&category).Error; err != nil {
		t.Fatalf("Failed to find category: %v", err)
	}
}

func TestUpsertBookIsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewImportRepository(db)

	// First call
	res1, err := repo.UpsertBook(sampleBook())
	if err != nil {
		t.Fatalf("First UpsertBook failed: %v", err)
	}

	if !res1.BookCreated {
		t.Error("Expected BookCreated to be true on first call")
	}

	// Second call with same book
	res2, err := repo.UpsertBook(sampleBook())
	if err != nil {
		t.Fatalf("Second UpsertBook failed: %v", err)
	}

	if res2.BookCreated {
		t.Error("Expected BookCreated to be false on second call")
	}

	if res2.CopiesAdded != 0 {
		t.Errorf("Expected CopiesAdded to be 0, got %d", res2.CopiesAdded)
	}

	if res2.CopiesSkipped != 3 {
		t.Errorf("Expected CopiesSkipped to be 3, got %d", res2.CopiesSkipped)
	}

	// Check database state is unchanged
	var book model.Book
	if err := db.Where("legacy_code = ?", 35).First(&book).Error; err != nil {
		t.Fatalf("Failed to find book: %v", err)
	}

	var copies []model.Copy
	if err := db.Where("book_id = ?", book.ID).Find(&copies).Error; err != nil {
		t.Fatalf("Failed to find copies: %v", err)
	}

	if len(copies) != 3 {
		t.Errorf("Expected 3 copies, got %d", len(copies))
	}
}

func TestUpsertBookEmptyValuesDoNotErase(t *testing.T) {
	db := setupTestDB(t)
	repo := NewImportRepository(db)

	// Insert sample book
	_, err := repo.UpsertBook(sampleBook())
	if err != nil {
		t.Fatalf("First UpsertBook failed: %v", err)
	}

	// Insert again with empty values for some fields
	bookWithEmptyValues := sampleBook()
	bookWithEmptyValues.UDK = ""
	bookWithEmptyValues.ISBN = ""
	bookWithEmptyValues.Language = ""
	bookWithEmptyValues.Pages = 300

	_, err = repo.UpsertBook(bookWithEmptyValues)
	if err != nil {
		t.Fatalf("Second UpsertBook failed: %v", err)
	}

	// Check that original values are preserved
	var book model.Book
	if err := db.Where("legacy_code = ?", 35).First(&book).Error; err != nil {
		t.Fatalf("Failed to find book: %v", err)
	}

	if book.UDK != "821" {
		t.Errorf("Expected UDK to be '821', got '%s'", book.UDK)
	}

	if book.ISBN != "9780306406157" {
		t.Errorf("Expected ISBN to be '9780306406157', got '%s'", book.ISBN)
	}

	if book.Language != "ukr" {
		t.Errorf("Expected Language to be 'ukr', got '%s'", book.Language)
	}

	if book.PageCount != 300 {
		t.Errorf("Expected PageCount to be 300, got %d", book.PageCount)
	}
}

func TestUpsertBookAuthorOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := NewImportRepository(db)

	// Insert sample book
	_, err := repo.UpsertBook(sampleBook())
	if err != nil {
		t.Fatalf("UpsertBook failed: %v", err)
	}

	// Find the book ID
	var book model.Book
	if err := db.Where("legacy_code = ?", 35).First(&book).Error; err != nil {
		t.Fatalf("Failed to find book: %v", err)
	}

	// Check author order
	rows, err := db.Raw("SELECT a.full_name FROM book_authors ba JOIN authors a ON a.id = ba.author_id WHERE ba.book_id = ? ORDER BY ba.position", book.ID).Rows()
	if err != nil {
		t.Fatalf("Failed to query authors: %v", err)
	}
	defer rows.Close()

	expectedAuthors := []string{"Zed Last", "Ann First"}
	i := 0
	for rows.Next() {
		var fullName string
		if err := db.ScanRows(rows, &fullName); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		if i >= len(expectedAuthors) {
			t.Errorf("Too many authors returned")
			break
		}
		if fullName != expectedAuthors[i] {
			t.Errorf("Expected author '%s', got '%s'", expectedAuthors[i], fullName)
		}
		i++
	}

	if i != len(expectedAuthors) {
		t.Errorf("Expected %d authors, got %d", len(expectedAuthors), i)
	}
}

func TestUpsertBookInventoryConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := NewImportRepository(db)

	// Insert sample book
	_, err := repo.UpsertBook(sampleBook())
	if err != nil {
		t.Fatalf("First UpsertBook failed: %v", err)
	}

	// Try to insert a second book with same inventory number
	bookWithSameInventory := model.ImportBook{
		LegacyCode:   36,
		Title:        "Another Book",
		Authors:      []string{"Test Author"},
		Year:         2000,
		City:         "Kyiv",
		Publisher:    "Pub",
		Department:   "1",
		Pages:        100,
		BBK:          "12.34",
		UDK:          "567",
		ISBN:         "9780306406158",
		Language:     "ukr",
		CategoryName: "Test Category",
		Inventory:    []string{"9001"}, // Same inventory number as first book
	}

	_, err = repo.UpsertBook(bookWithSameInventory)
	if err == nil {
		t.Fatal("Expected error for inventory conflict")
	}

	if !errors.Is(err, ErrConflict) {
		t.Errorf("Expected ErrConflict, got %v", err)
	}

	// Check that second book was not inserted
	var book model.Book
	err = db.Where("legacy_code = ?", 36).First(&book).Error
	if err == nil {
		t.Error("Second book should not have been inserted")
	}
}

func TestUpsertBookISBNConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := NewImportRepository(db)

	// Insert sample book
	_, err := repo.UpsertBook(sampleBook())
	if err != nil {
		t.Fatalf("First UpsertBook failed: %v", err)
	}

	// Try to insert a second book with same ISBN
	bookWithSameISBN := model.ImportBook{
		LegacyCode:   36,
		Title:        "Another Book",
		Authors:      []string{"Test Author"},
		Year:         2000,
		City:         "Kyiv",
		Publisher:    "Pub",
		Department:   "1",
		Pages:        100,
		BBK:          "12.34",
		UDK:          "567",
		ISBN:         "9780306406157", // Same ISBN as first book
		Language:     "ukr",
		CategoryName: "Test Category",
		Inventory:    []string{"9100"},
	}

	_, err = repo.UpsertBook(bookWithSameISBN)
	if err == nil {
		t.Fatal("Expected error for ISBN conflict")
	}

	if !errors.Is(err, ErrConflict) {
		t.Errorf("Expected ErrConflict, got %v", err)
	}
}

func TestListExportRowsGroupsByDepartment(t *testing.T) {
	db := setupTestDB(t)
	repo := NewImportRepository(db)

	// Insert sample book with department "3"
	_, err := repo.UpsertBook(sampleBook())
	if err != nil {
		t.Fatalf("First UpsertBook failed: %v", err)
	}

	// Insert same book with different department
	bookWithDifferentDepartment := sampleBook()
	bookWithDifferentDepartment.Department = "5"
	bookWithDifferentDepartment.Inventory = []string{"9100"}

	_, err = repo.UpsertBook(bookWithDifferentDepartment)
	if err != nil {
		t.Fatalf("Second UpsertBook failed: %v", err)
	}

	// Get export rows
	exportRepo := NewExportRepository(db)
	rows, err := exportRepo.ListExportRows()
	if err != nil {
		t.Fatalf("ListExportRows failed: %v", err)
	}

	// Should have 2 rows for this book
	var bookRows []model.ExportRow
	for _, row := range rows {
		if row.LegacyCode != nil && *row.LegacyCode == 35 {
			bookRows = append(bookRows, row)
		}
	}

	if len(bookRows) != 2 {
		t.Errorf("Expected 2 rows for book with legacy code 35, got %d", len(bookRows))
	}

	// Check that we have one row for department "3" and one for department "5"
	dept3Found := false
	dept5Found := false

	for _, row := range bookRows {
		if row.Department == "3" {
			dept3Found = true
			if len(row.Inventory) != 3 {
				t.Errorf("Expected 3 inventory numbers for department '3', got %d", len(row.Inventory))
			}
			if row.Authors != "Zed Last; Ann First" {
				t.Errorf("Expected authors 'Zed Last; Ann First', got '%s'", row.Authors)
			}
		} else if row.Department == "5" {
			dept5Found = true
			if len(row.Inventory) != 1 {
				t.Errorf("Expected 1 inventory number for department '5', got %d", len(row.Inventory))
			}
			if row.Authors != "Zed Last; Ann First" {
				t.Errorf("Expected authors 'Zed Last; Ann First', got '%s'", row.Authors)
			}
		}
	}

	if !dept3Found {
		t.Error("Row for department '3' not found")
	}

	if !dept5Found {
		t.Error("Row for department '5' not found")
	}
}
