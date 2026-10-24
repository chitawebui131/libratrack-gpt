package repository

import (
	"errors"
	"testing"
	"time"

	"libratrack/internal/model"
)

func TestReturnOK(t *testing.T) {
	// Seed 1 copy and 1 user, issue a loan. Return it with a calcFine stub that always returns 6.
	// Result: ReturnedAt is not nil, FineAmount is 6, and the copy status in the database is "available" again.
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

	returnedAt := time.Now()
	fine := 6
	loan, err = loanRepo.Return(loan.ID, returnedAt, func(dueAt, returnedAt time.Time) int {
		return fine
	})
	if err != nil {
		t.Fatal(err)
	}

	if loan.ReturnedAt == nil {
		t.Fatal("ReturnedAt must not be nil")
	}

	if loan.FineAmount != fine {
		t.Fatalf("Expected FineAmount %d, got %d", fine, loan.FineAmount)
	}

	var copy model.Copy
	err = db.First(&copy, copyIDs[0]).Error
	if err != nil {
		t.Fatal(err)
	}

	if copy.Status != "available" {
		t.Fatalf("expected copy status 'available', got '%s'", copy.Status)
	}
}

func TestReturnTwice(t *testing.T) {
	// Return the same loan twice. The second call returns ErrAlreadyReturned.
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

	returnedAt := time.Now()
	fine := 6
	_, err = loanRepo.Return(loan.ID, returnedAt, func(dueAt, returnedAt time.Time) int {
		return fine
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = loanRepo.Return(loan.ID, returnedAt, func(dueAt, returnedAt time.Time) int {
		return fine
	})
	if !errors.Is(err, ErrAlreadyReturned) {
		t.Fatalf("expected ErrAlreadyReturned, got %v", err)
	}
}

func TestReturnUnknownLoan(t *testing.T) {
	// Return loan id 9999 -> ErrNotFound.
	db := setupTestDB(t)
	seedLoanFixtures(t, db, 0, 0)

	loanRepo := NewLoanRepository(db)
	returnedAt := time.Now()
	fine := 6
	_, err := loanRepo.Return(9999, returnedAt, func(dueAt, returnedAt time.Time) int {
		return fine
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
