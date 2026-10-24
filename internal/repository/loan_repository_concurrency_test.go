package repository

import (
	"errors" // 1. Dodano import pakietu
	"sync"
	"testing"
	"time"
)

func TestIssueConcurrent(t *testing.T) {
	db := setupTestDB(t)
	copyIDs, userIDs := seedLoanFixtures(t, db, 1, 5)

	repo := &gormLoanRepository{db: db}
	now := time.Now()

	var wg sync.WaitGroup
	errs := make([]error, 5) // 2. Zmieniono nazwe z 'errors' na 'errs'

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = repo.Issue(copyIDs[0], userIDs[i], now, now.Add(14*24*time.Hour), 5)
		}(i)
	}

	wg.Wait()

	nilCount := 0
	for _, err := range errs { // 2. Zmieniono na 'errs'
		if err == nil {
			nilCount++
		} else if !errors.Is(err, ErrCopyNotAvailable) && !errors.Is(err, ErrConflict) { // Teraz kompilator prawidlowo uzywa pakietu
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if nilCount != 1 {
		t.Fatalf("expected exactly one nil error, got %d", nilCount)
	}
}
