package repository

import (
	"testing"
	"time"

	"libratrack/internal/model"
)

func TestListByStatus(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 3, 1)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)

	// Loan A: Issue, then Return (a returned loan).
	loanA, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAt, 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = loanRepo.Return(loanA.ID, time.Now(), func(dueAt time.Time, returnedAt time.Time) int {
		return 0
	})
	if err != nil {
		t.Fatal(err)
	}

	// Loan C: insert directly with db.Create(&model.Loan{...}) with DueAt in the past, BorrowedAt before DueAt, ReturnedAt nil (an overdue loan).
	loanC := model.Loan{
		CopyID:     copyIDs[2],
		UserID:     userIDs[0],
		BorrowedAt: borrowedAt.Add(-30 * 24 * time.Hour),
		DueAt:      borrowedAt.Add(-1 * 24 * time.Hour),
		ReturnedAt: nil,
	}
	if err := db.Create(&loanC).Error; err != nil {
		t.Fatal(err)
	}

	// List("active") returns B and C. List("returned") returns A. List("overdue") returns only C. List("") returns all three.
	loans, err := loanRepo.List("active", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(loans) != 2 {
		t.Fatalf("expected 2 active loans, got %d", len(loans))
	}

	loans, err = loanRepo.List("returned", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(loans) != 1 {
		t.Fatalf("expected 1 returned loan, got %d", len(loans))
	}

	loans, err = loanRepo.List("overdue", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(loans) != 1 {
		t.Fatalf("expected 1 overdue loan, got %d", len(loans))
	}

	loans, err = loanRepo.List("", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(loans) != 3 {
		t.Fatalf("expected 3 loans, got %d", len(loans))
	}
}

func TestListByUserPagination(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 3, 2)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)

	// Give user 1 three loans and user 2 one loan.
	_, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAt, 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = loanRepo.Issue(copyIDs[1], userIDs[0], borrowedAt, dueAt, 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = loanRepo.Issue(copyIDs[2], userIDs[0], borrowedAt, dueAt, 5)
	if err != nil {
		t.Fatal(err)
	}
	_, err = loanRepo.Issue(copyIDs[0], userIDs[1], borrowedAt, dueAt, 5)
	if err != nil {
		t.Fatal(err)
	}

	// ListByUser(user1, 1, 2) returns 2 loans. ListByUser(user1, 2, 2) returns 1 loan. ListByUser(user2, 1, 10) returns 1 loan.
	loans, err := loanRepo.ListByUser(userIDs[0], 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(loans) != 2 {
		t.Fatalf("expected 2 loans for user 1, page 1, got %d", len(loans))
	}

	loans, err = loanRepo.ListByUser(userIDs[0], 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(loans) != 1 {
		t.Fatalf("expected 1 loan for user 1, page 2, got %d", len(loans))
	}

	loans, err = loanRepo.ListByUser(userIDs[1], 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(loans) != 1 {
		t.Fatalf("expected 1 loan for user 2, got %d", len(loans))
	}
}
