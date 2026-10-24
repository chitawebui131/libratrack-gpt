package repository

import (
	"testing"
	"time"
)

func TestListViewsByUserOnlyOwnLoans(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 3, 2)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)
	maxActiveLoans := 5

	// Issue 2 loans for user 1
	_, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}
	_, err = loanRepo.Issue(copyIDs[1], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	// Issue 1 loan for user 2
	_, err = loanRepo.Issue(copyIDs[2], userIDs[1], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	viewRepo := NewLoanViewRepository(db)

	rows, total, err := viewRepo.ListViewsByUser(userIDs[0], 1, 20)
	if err != nil {
		t.Fatal(err)
	}

	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	for _, row := range rows {
		if row.UserID != userIDs[0] {
			t.Errorf("expected user ID %d, got %d", userIDs[0], row.UserID)
		}
		if row.BookTitle == "" {
			t.Error("expected non-empty BookTitle")
		}
		if row.InventoryNo == "" {
			t.Error("expected non-empty InventoryNo")
		}
	}
}

func TestListActiveViewsOrderedByDueDate(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 3, 1)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now()
	maxActiveLoans := 5

	// Loan A: due in 10 days
	dueAtA := borrowedAt.Add(10 * 24 * time.Hour)
	_, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAtA, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	// Loan B: due in 3 days
	dueAtB := borrowedAt.Add(3 * 24 * time.Hour)
	_, err = loanRepo.Issue(copyIDs[1], userIDs[0], borrowedAt, dueAtB, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	// Loan C: due in 20 days
	dueAtC := borrowedAt.Add(20 * 24 * time.Hour)
	_, err = loanRepo.Issue(copyIDs[2], userIDs[0], borrowedAt, dueAtC, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	viewRepo := NewLoanViewRepository(db)

	rows, total, err := viewRepo.ListActiveViews(1, 20)
	if err != nil {
		t.Fatal(err)
	}

	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	// Check that the loans are ordered by due date (earliest first)
	if rows[0].DueAt.After(rows[1].DueAt) {
		t.Error("expected first row to have earlier due date than second")
	}
	if rows[1].DueAt.After(rows[2].DueAt) {
		t.Error("expected second row to have earlier due date than third")
	}

	// Return loan C
	_, err = loanRepo.Return(copyIDs[2], time.Now(), func(dueAt time.Time, returnedAt time.Time) int {
		return 0
	})
	if err != nil {
		t.Fatal(err)
	}

	rows, total, err = viewRepo.ListActiveViews(1, 20)
	if err != nil {
		t.Fatal(err)
	}

	if total != 2 {
		t.Fatalf("expected total 2 after returning loan C, got %d", total)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows after returning loan C, got %d", len(rows))
	}

	// Check that the loans are ordered by due date (earliest first)
	if rows[0].DueAt.After(rows[1].DueAt) {
		t.Error("expected first row to have earlier due date than second")
	}
}

func TestTopBooksOrderAndExclusion(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 3, 1)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now()
	dueAt := borrowedAt.Add(14 * 24 * time.Hour)
	maxActiveLoans := 5

	// Issue 2 loans for copy 0 (same book)
	_, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}
	_, err = loanRepo.Issue(copyIDs[1], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	// Issue 1 loan for copy 2 (different book)
	_, err = loanRepo.Issue(copyIDs[2], userIDs[0], borrowedAt, dueAt, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	reportRepo := NewReportRepository(db)

	topBooks, err := reportRepo.TopBooks(10)
	if err != nil {
		t.Fatal(err)
	}

	if len(topBooks) < 2 {
		t.Fatalf("expected at least 2 books, got %d", len(topBooks))
	}

	// First book should be the one with 2 loans
	if topBooks[0].LoanCount != 2 {
		t.Errorf("expected first book to have 2 loans, got %d", topBooks[0].LoanCount)
	}

	// Second book should be the one with 1 loan
	if topBooks[1].LoanCount != 1 {
		t.Errorf("expected second book to have 1 loan, got %d", topBooks[1].LoanCount)
	}
}

func TestOverdueLoansAndStats(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 3, 1)

	loanRepo := NewLoanRepository(db)
	borrowedAt := time.Now().Add(-20 * 24 * time.Hour) // 20 days ago
	maxActiveLoans := 5

	// Loan A: due in the past (overdue)
	dueAtA := borrowedAt.Add(14 * 24 * time.Hour) // 6 days ago
	_, err := loanRepo.Issue(copyIDs[0], userIDs[0], borrowedAt, dueAtA, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	// Loan B: due in the future
	dueAtB := borrowedAt.Add(30 * 24 * time.Hour) // 10 days in the future
	_, err = loanRepo.Issue(copyIDs[1], userIDs[0], borrowedAt, dueAtB, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	// Loan C: past due but returned
	dueAtC := borrowedAt.Add(7 * 24 * time.Hour) // 13 days ago
	_, err = loanRepo.Issue(copyIDs[2], userIDs[0], borrowedAt, dueAtC, maxActiveLoans)
	if err != nil {
		t.Fatal(err)
	}

	// Return loan C
	_, err = loanRepo.Return(copyIDs[2], time.Now(), func(dueAt time.Time, returnedAt time.Time) int {
		return 0
	})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	reportRepo := NewReportRepository(db)

	overdueLoans, err := reportRepo.OverdueLoans(now)
	if err != nil {
		t.Fatal(err)
	}

	if len(overdueLoans) != 1 {
		t.Fatalf("expected 1 overdue loan, got %d", len(overdueLoans))
	}

	if overdueLoans[0].LoanID != copyIDs[0] {
		t.Error("expected overdue loan to be the first one")
	}

	stats, err := reportRepo.Stats(now)
	if err != nil {
		t.Fatal(err)
	}

	if stats.ActiveLoans != 2 {
		t.Errorf("expected 2 active loans, got %d", stats.ActiveLoans)
	}

	if stats.OverdueLoans != 1 {
		t.Errorf("expected 1 overdue loan, got %d", stats.OverdueLoans)
	}

	if stats.TotalCopies != 3 {
		t.Errorf("expected 3 total copies, got %d", stats.TotalCopies)
	}
}
