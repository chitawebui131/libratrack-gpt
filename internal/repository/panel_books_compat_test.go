package repository

import (
	"testing"

	"libratrack/internal/model"
)

func TestSaveBookStoresUDKLanguageAndAuthorOrder(t *testing.T) {
	// 1. db := setupTestDB(t). Create one category and three authors "Zed Last", "Ann First", "Mid Middle" with db.Create. Fatal on error.
	db := setupTestDB(t)
	category := mustCategory(t, db, "Test Category")
	zed := mustAuthor(t, db, "Zed Last")
	ann := mustAuthor(t, db, "Ann First")
	mid := mustAuthor(t, db, "Mid Middle")

	// 2. repo := NewPanelRepository(db). b := &model.Book{Title: "Ordered", CategoryID: <category id>, UDK: "821.161.2(477)", Language: "ukr"}.
	repo := NewPanelRepository(db)
	b := &model.Book{
		Title:         "Ordered",
		CategoryID:    category.ID,
		UDK:           "821.161.2(477)",
		Language:      "ukr",
		Publisher:     "Test Publisher",
		PublishedYear: 2023,
		Description:   "Test description",
	}

	// 3. err := repo.SaveBook(b, []uint{<Zed id>, <Ann id>, <Mid id>}). Fatal on error.
	err := repo.SaveBook(b, []uint{zed.ID, ann.ID, mid.ID})
	if err != nil {
		t.Fatalf("SaveBook failed: %v", err)
	}

	// 4. Read the book back with db.First. UDK and Language are stored.
	var savedBook model.Book
	err = db.First(&savedBook, b.ID).Error
	if err != nil {
		t.Fatalf("Failed to read book back: %v", err)
	}
	if savedBook.UDK != "821.161.2(477)" {
		t.Errorf("UDK not stored correctly, got %s, want 821.161.2(477)", savedBook.UDK)
	}
	if savedBook.Language != "ukr" {
		t.Errorf("Language not stored correctly, got %s, want ukr", savedBook.Language)
	}

	// 5. Read the positions with db.Raw("SELECT author_id, position FROM book_authors WHERE book_id = ? ORDER BY position", b.ID).Scan(&rows) (rows is a slice of a small struct with AuthorID and Position). The order of author ids must be Zed, Ann, Mid with positions 0, 1, 2 (NOT alphabetical).
	type bookAuthor struct {
		AuthorID uint
		Position uint
	}
	var rows []bookAuthor
	err = db.Raw("SELECT author_id, position FROM book_authors WHERE book_id = ? ORDER BY position", b.ID).Scan(&rows).Error
	if err != nil {
		t.Fatalf("Failed to read author positions: %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("Expected 3 authors, got %d", len(rows))
	}

	expectedOrder := []uint{zed.ID, ann.ID, mid.ID}
	for i, row := range rows {
		if row.AuthorID != expectedOrder[i] {
			t.Errorf("Author order incorrect at position %d: got %d, want %d", i, row.AuthorID, expectedOrder[i])
		}
		if row.Position != uint(i) {
			t.Errorf("Position incorrect at author %d: got %d, want %d", i, row.Position, i)
		}
	}
}

func TestSaveBookUpdateKeepsCompatFields(t *testing.T) {
	// 1. Create a book with SaveBook and UDK "1", Language "eng". Then insert legacy data directly: db.Model(&model.Book{}).Where("id = ?", b.ID).Updates(map[string]interface{}{"bbk": "34.61", "page_count": 100}).
	db := setupTestDB(t)
	category := mustCategory(t, db, "Test Category")
	author := mustAuthor(t, db, "Test Author")

	repo := NewPanelRepository(db)
	b := &model.Book{
		Title:         "Test Book",
		CategoryID:    category.ID,
		UDK:           "1",
		Language:      "eng",
		Publisher:     "Test Publisher",
		PublishedYear: 2023,
		Description:   "Test description",
	}
	err := repo.SaveBook(b, []uint{author.ID})
	if err != nil {
		t.Fatalf("SaveBook failed: %v", err)
	}

	// Insert legacy data directly
	err = db.Model(&model.Book{}).Where("id = ?", b.ID).Updates(map[string]interface{}{
		"bbk":        "34.61",
		"page_count": 100,
	}).Error
	if err != nil {
		t.Fatalf("Failed to insert legacy data: %v", err)
	}

	// 2. Call SaveBook again with the same ID and a new title (UDK "2", Language "ukr").
	b.Title = "Updated Title"
	b.UDK = "2"
	b.Language = "ukr"
	err = repo.SaveBook(b, []uint{author.ID})
	if err != nil {
		t.Fatalf("SaveBook failed on update: %v", err)
	}

	// 3. The book has the new title, UDK "2", Language "ukr", AND bbk "34.61" and page_count 100 are unchanged (SaveBook must not wipe fields it does not know about).
	var updatedBook model.Book
	err = db.First(&updatedBook, b.ID).Error
	if err != nil {
		t.Fatalf("Failed to read updated book: %v", err)
	}

	if updatedBook.Title != "Updated Title" {
		t.Errorf("Title not updated correctly, got %s, want Updated Title", updatedBook.Title)
	}
	if updatedBook.UDK != "2" {
		t.Errorf("UDK not updated correctly, got %s, want 2", updatedBook.UDK)
	}
	if updatedBook.Language != "ukr" {
		t.Errorf("Language not updated correctly, got %s, want ukr", updatedBook.Language)
	}
	if updatedBook.BBK != "34.61" {
		t.Errorf("BBK not preserved correctly, got %s, want 34.61", updatedBook.BBK)
	}
	if updatedBook.PageCount != 100 {
		t.Errorf("Page count not preserved correctly, got %d, want 100", updatedBook.PageCount)
	}
}
