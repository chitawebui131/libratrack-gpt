package service

import "time"

// CalculateFine returns the fine in whole UAH. Rules:
//  1. Only the UTC calendar dates of dueAt and returnedAt count.
//  2. overdueDays = whole calendar days from the due date to the return date; 0 if the return date is not after the due date.
//  3. fine = overdueDays * FinePerDay, but never more than FineMax.
//  4. The result is an integer. No fractions, no float.
func (p LoanPolicy) CalculateFine(dueAt, returnedAt time.Time) int {
	// 1. due := dateOnly(dueAt) and ret := dateOnly(returnedAt).
	due := dateOnly(dueAt)
	ret := dateOnly(returnedAt)
	// 2. If !ret.After(due), return 0.
	if !ret.After(due) {
		return 0
	}
	// 3. days := int(ret.Sub(due) / (24 * time.Hour)).
	days := int(ret.Sub(due) / (24 * time.Hour))
	// 4. fine := days * p.FinePerDay. If fine > p.FineMax, set fine = p.FineMax.
	fine := days * p.FinePerDay
	if fine > p.FineMax {
		fine = p.FineMax
	}
	// 5. Return fine.
	return fine
}

// dateOnly returns midnight UTC of the date of t.
func dateOnly(t time.Time) time.Time {
	// u := t.UTC(); return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC).
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}
