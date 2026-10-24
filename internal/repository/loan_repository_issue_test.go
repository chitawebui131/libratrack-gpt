package repository

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

// seedLoanFixtures creates one book, `copies` copies with status available, and `users` users. It returns their ids.
func seedLoanFixtures(t *testing.T, db *gorm.DB, copies, users int) (copyIDs, userIDs []uint) {
	t.Helper()
	category := mustCategory(t, db, "Test Category")
	author := mustAuthor(t, db, "Test Author")
	book := mustBook(t, db, "Test Book", "1234567890", category.ID, []uint{author.ID})

	copyIDs = make([]uint, copies)
	for i := 0; i < copies; i++ {
		copy := model.Copy{
			BookID:      book.ID,
			Status:      model.CopyStatusAvailable,
			InventoryNo: "INV0000000000" + string(rune(i+'0')),
		}
		if err := db.Create(&copy).Error; err != nil {
			t.Fatal(err)
		}
		copyIDs[i] = copy.ID
	}

	userIDs = make([]uint, users)
	for i := 0; i < users; i++ {
		user := model.User{
			Email:        "user" + string(rune(i+'0')) + "@example.com",
			PasswordHash: "hash",
			FullName:     "User " + string(rune(i+'0')),
			Role:         "reader",
		}
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
		userIDs[i] = user.ID
	}

	return copyIDs, userIDs
}

func TestIssueOK(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 1, 1)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)
	maxActiveLoans := 5

	loan, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	if loan.Copy == nil || loan.Copy.Book == nil {
		t.Fatal("loan must have Copy.Book loaded")
	}

	var copy model.Copy
	err = db.First(&copy, copyIDs[0]).Error
	if err != nil {
		t.Fatal(err)
	}

	if copy.Status != "borrowed" {
		t.Fatalf("expected copy status 'borrowed', got '%s'", copy.Status)
	}
}

func TestIssueSameCopyTwice(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 1, 2)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)
	maxActiveLoans := 5

	_, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	_, err = loanRepo.Issue(copyIDs[0], userIDs[1], borrowedAt, dueAt, maxActiveLoans)
	if !errors.Is(err, ErrCopyNotAvailable) {
		t.Fatalf("expected ErrCopyNotAvailable, got %v", err)
	}
}

func TestIssueUnknownReferences(t *testing.T) {
	db := setupTestDB(t)
	_, userIDs := seedLoanFixtures(t, db, 0, 1)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)
	maxActiveLoans := 5

	_, err := loanRepo.Issue(9999, userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}

	_, err = loanRepo.Issue(1, userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if !errors.Is(err, ErrCopyNotAvailable) {
		t.Fatalf("expected ErrCopyNotAvailable, got %v", err)
	}
}

func TestIssueLimit(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 3, 1)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)
	maxActiveLoans := 2

	_, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	_, err = loanRepo.Issue(copyIDs[1], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	_, err = loanRepo.Issue(copyIDs[2], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if !errors.Is(err, ErrLoanLimitReached) {
		t.Fatalf("expected ErrLoanLimitReached, got %v", err)
	}
}
