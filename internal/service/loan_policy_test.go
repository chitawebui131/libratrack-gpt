package service

import (
	"testing"
	"time"
)

func utc(y int, m time.Month, d, h, min int) time.Time {
	return time.Date(y, m, d, h, min, 0, 0, time.UTC)
}

func testPolicy(t *testing.T) LoanPolicy {
	p, err := NewLoanPolicy(14, 2, 100, 5)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDueDate(t *testing.T) {
	p := testPolicy(t)
	borrowed := utc(2025, 1, 1, 10, 0)
	due := p.DueDate(borrowed)
	expected := utc(2025, 1, 15, 10, 0)
	if !due.Equal(expected) {
		t.Errorf("DueDate: got %v, want %v", due, expected)
	}
}

func TestCalculateFine(t *testing.T) {
	p := testPolicy(t)
	due := utc(2025, 1, 15, 12, 0)
	tests := []struct {
		returned time.Time
		want     int
	}{
		{utc(2025, 1, 10, 9, 0), 0},
		{utc(2025, 1, 15, 12, 0), 0},
		{utc(2025, 1, 15, 23, 30), 0},
		{utc(2025, 1, 16, 0, 30), 2},
		{utc(2025, 1, 20, 8, 0), 10},
		{utc(2025, 3, 6, 12, 0), 100},
		{utc(2025, 6, 1, 12, 0), 100},
	}
	for _, tc := range tests {
		got := p.CalculateFine(due, tc.returned)
		if got != tc.want {
			t.Errorf("CalculateFine(%v, %v): got %d, want %d", due, tc.returned, got, tc.want)
		}
	}
}

func TestNewLoanPolicyRejectsBadValues(t *testing.T) {
	tests := []struct {
		periodDays     int
		finePerDay     int
		fineMax        int
		maxActiveLoans int
	}{
		{0, 2, 100, 5},
		{14, 0, 100, 5},
		{14, 2, 0, 5},
		{14, 2, 100, 0},
		{14, 5, 2, 5},
	}
	for _, tc := range tests {
		_, err := NewLoanPolicy(tc.periodDays, tc.finePerDay, tc.fineMax, tc.maxActiveLoans)
		if err == nil {
			t.Errorf("NewLoanPolicy(%d, %d, %d, %d): expected error, got nil", tc.periodDays, tc.finePerDay, tc.fineMax, tc.maxActiveLoans)
		}
	}
}
