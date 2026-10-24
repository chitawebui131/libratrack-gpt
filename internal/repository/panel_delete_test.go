package repository

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"libratrack/internal/model"
)

func TestTranslateErrorPgx(t *testing.T) {
	err := translateError(fmt.Errorf("wrap: %w", &pgconn.PgError{Code: "23505"}))
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	err = translateError(fmt.Errorf("wrap: %w", &pgconn.PgError{Code: "23503"}))
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}

	err = translateError(errors.New("x"))
	if err.Error() != "x" {
		t.Fatalf("expected \"x\", got %v", err)
	}
}

func TestDeleteBookWithCopiesWithoutLoans(t *testing.T) {
	db := setupTestDB(t)
	category := mustCategory(t, db, "Test Category")
	author := mustAuthor(t, db, "Test Author")
	book := mustBook(t, db, "Test Book", "1234567890", category.ID, []uint{author.ID})

	cp1 := model.Copy{
		BookID:      book.ID,
		Status:      model.CopyStatusAvailable,
		InventoryNo: "INV00000000001",
		Department:  "3",
	}
	cp2 := model.Copy{
		BookID:      book.ID,
		Status:      model.CopyStatusAvailable,
		InventoryNo: "INV00000000002",
		Department:  "4",
	}
	if err := db.Create(&cp1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&cp2).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewPanelRepository(db)
	err := repo.DeleteBook(book.ID)
	if err != nil {
		t.Fatal(err)
	}

	var b model.Book
	err = db.First(&b, book.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	var c1 model.Copy
	err = db.First(&c1, cp1.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
	var c2 model.Copy
	err = db.First(&c2, cp2.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	var ba []model.BookAuthor
	err = db.Find(&ba, "book_id = ?", book.ID).Error
	if err != nil {
		t.Fatal(err)
	}
	if len(ba) != 0 {
		t.Fatalf("expected no book_authors rows, got %d", len(ba))
	}

	var a model.Author
	err = db.First(&a, author.ID).Error
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteBookWithLoanHistoryIsBlocked(t *testing.T) {
	db := setupTestDB(t)
	category := mustCategory(t, db, "Test Category")
	author := mustAuthor(t, db, "Test Author")
	book := mustBook(t, db, "Test Book", "1234567890", category.ID, []uint{author.ID})

	cp := model.Copy{
		BookID:      book.ID,
		Status:      model.CopyStatusAvailable,
		InventoryNo: "INV00000000001",
	}
	if err := db.Create(&cp).Error; err != nil {
		t.Fatal(err)
	}

	user := model.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		FullName:     "User",
		Role:         "reader",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	loan := model.Loan{
		CopyID:     cp.ID,
		UserID:     user.ID,
		BorrowedAt: db.NowFunc(),
		DueAt:      db.NowFunc().Add(14 * 24 * time.Hour),
	}
	if err := db.Create(&loan).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewPanelRepository(db)
	err := repo.DeleteBook(book.ID)
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}

	var b model.Book
	err = db.First(&b, book.ID).Error
	if err != nil {
		t.Fatal(err)
	}

	var c model.Copy
	err = db.First(&c, cp.ID).Error
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteBookNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPanelRepository(db)
	err := repo.DeleteBook(999999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteAuthorAndCategoryProtected(t *testing.T) {
	db := setupTestDB(t)
	category := mustCategory(t, db, "Test Category")
	author := mustAuthor(t, db, "Test Author")
	_ = mustBook(t, db, "Test Book", "1234567890", category.ID, []uint{author.ID})

	dictRepo := NewDictionaryRepository(db)

	err := dictRepo.DeleteAuthor(author.ID)
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}

	err = dictRepo.DeleteCategory(category.ID)
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}

	unusedAuthor := mustAuthor(t, db, "Unused Author")
	unusedCategory := mustCategory(t, db, "Unused Category")

	err = dictRepo.DeleteAuthor(unusedAuthor.ID)
	if err != nil {
		t.Fatal(err)
	}

	var a model.Author
	err = db.First(&a, unusedAuthor.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	err = dictRepo.DeleteCategory(unusedCategory.ID)
	if err != nil {
		t.Fatal(err)
	}

	var c model.Category
	err = db.First(&c, unusedCategory.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestCopyUpdateAndDelete(t *testing.T) {
	db := setupTestDB(t)
	category := mustCategory(t, db, "Test Category")
	author := mustAuthor(t, db, "Test Author")
	book := mustBook(t, db, "Test Book", "1234567890", category.ID, []uint{author.ID})

	repo := NewPanelRepository(db)

	err := repo.AddCopy(book.ID, "A-1", "3")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.AddCopy(book.ID, "A-2", "")
	if err != nil {
		t.Fatal(err)
	}

	var copies []model.Copy
	if err := db.Order("id ASC").Where("book_id = ?", book.ID).Find(&copies).Error; err != nil {
		t.Fatal(err)
	}
	if len(copies) < 2 {
		t.Fatalf("expected 2 copies, got %d", len(copies))
	}
	cp1ID := copies[0].ID
	cp2ID := copies[1].ID

	err = repo.UpdateCopy(cp1ID, "A-2", "")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	err = repo.UpdateCopy(cp1ID, "A-9", "5")
	if err != nil {
		t.Fatal(err)
	}

	var cp model.Copy
	err = db.First(&cp, cp1ID).Error
	if err != nil {
		t.Fatal(err)
	}
	if cp.InventoryNo != "A-9" {
		t.Fatalf("expected inventory_no 'A-9', got '%s'", cp.InventoryNo)
	}
	if cp.Department != "5" {
		t.Fatalf("expected department '5', got '%s'", cp.Department)
	}

	user := model.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		FullName:     "User",
		Role:         "reader",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	loan := model.Loan{
		CopyID:     cp2ID,
		UserID:     user.ID,
		BorrowedAt: db.NowFunc(),
		DueAt:      db.NowFunc().Add(14 * 24 * time.Hour),
	}
	if err := db.Create(&loan).Error; err != nil {
		t.Fatal(err)
	}

	err = repo.DeleteCopy(cp2ID)
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}

	err = repo.DeleteCopy(cp1ID)
	if err != nil {
		t.Fatal(err)
	}

	var c model.Copy
	err = db.First(&c, cp1ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	err = repo.DeleteCopy(999999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
