package repository

import (
	"errors"
	"testing"
	"time"

	"libratrack/internal/model"
)

func TestUpdateCopyStatusAllowed(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, _ := seedLoanFixtures(t, db, 1, 1)

	panelRepo := NewPanelRepository(db)

	err := panelRepo.UpdateCopyStatus(copyIDs[0], "lost")
	if err != nil {
		t.Fatal(err)
	}

	var copy model.Copy
	err = db.First(&copy, copyIDs[0]).Error
	if err != nil {
		t.Fatal(err)
	}

	if copy.Status != "lost" {
		t.Fatalf("expected copy status 'lost', got '%s'", copy.Status)
	}

	err = panelRepo.UpdateCopyStatus(copyIDs[0], "available")
	if err != nil {
		t.Fatal(err)
	}

	err = db.First(&copy, copyIDs[0]).Error
	if err != nil {
		t.Fatal(err)
	}

	if copy.Status != "available" {
		t.Fatalf("expected copy status 'available', got '%s'", copy.Status)
	}
}

func TestUpdateCopyStatusRejectsBorrowed(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 1, 1)

	panelRepo := NewPanelRepository(db)
	loanRepo := NewLoanRepository(db)

	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)
	maxActiveLoans := 5

	_, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	err = panelRepo.UpdateCopyStatus(copyIDs[0], "borrowed")
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}
}

func TestUpdateCopyStatusActiveLoan(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 1, 1)

	panelRepo := NewPanelRepository(db)
	loanRepo := NewLoanRepository(db)

	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)
	maxActiveLoans := 5

	_, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	err = panelRepo.UpdateCopyStatus(copyIDs[0], "lost")
	if !errors.Is(err, ErrCopyNotAvailable) {
		t.Fatalf("expected ErrCopyNotAvailable, got %v", err)
	}
}

func TestUpdateCopyStatusUnknownCopy(t *testing.T) {
	db := setupTestDB(t)

	panelRepo := NewPanelRepository(db)

	err := panelRepo.UpdateCopyStatus(9999, "lost")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAddCopy(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, _ := seedLoanFixtures(t, db, 1, 0)

	panelRepo := NewPanelRepository(db)

	var copy model.Copy
	err := db.First(&copy, copyIDs[0]).Error
	if err != nil {
		t.Fatal(err)
	}

	bookID := copy.BookID

	err = panelRepo.AddCopy(bookID, "INV-NEW-1", "")
	if err != nil {
		t.Fatal(err)
	}

	err = panelRepo.AddCopy(bookID, "INV-NEW-1", "")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	err = panelRepo.AddCopy(9999, "X", "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
