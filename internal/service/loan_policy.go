package service

import (
	"errors"
	"time"
)

type LoanPolicy struct {
	PeriodDays     int
	FinePerDay     int
	FineMax        int
	MaxActiveLoans int
}

func NewLoanPolicy(periodDays, finePerDay, fineMax, maxActiveLoans int) (LoanPolicy, error) {
	// 1. If any of the four values is <= 0, return LoanPolicy{} and errors.New("loan policy values must be positive").
	if periodDays <= 0 || finePerDay <= 0 || fineMax <= 0 || maxActiveLoans <= 0 {
		return LoanPolicy{}, errors.New("loan policy values must be positive")
	}
	// 2. If fineMax < finePerDay, return LoanPolicy{} and errors.New("fine max must not be less than fine per day").
	if fineMax < finePerDay {
		return LoanPolicy{}, errors.New("fine max must not be less than fine per day")
	}
	// 3. Return LoanPolicy{PeriodDays: periodDays, FinePerDay: finePerDay, FineMax: fineMax, MaxActiveLoans: maxActiveLoans}, nil.
	return LoanPolicy{PeriodDays: periodDays, FinePerDay: finePerDay, FineMax: fineMax, MaxActiveLoans: maxActiveLoans}, nil
}

// DueDate returns borrowedAt (in UTC) plus PeriodDays * 24 hours.
func (p LoanPolicy) DueDate(borrowedAt time.Time) time.Time {
	// Return borrowedAt.UTC().Add(time.Duration(p.PeriodDays) * 24 * time.Hour).
	return borrowedAt.UTC().Add(time.Duration(p.PeriodDays) * 24 * time.Hour)
}
