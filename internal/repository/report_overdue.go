package repository

import (
	"time"

	"libratrack/internal/model"
)

// OverdueLoans returns active loans with due_at before now, ordered by due date.
func (r *gormReportRepository) OverdueLoans(now time.Time) ([]model.LoanView, error) {
	// 1. var rows []model.LoanView.
	var rows []model.LoanView
	// 2. err := r.db.Raw(loanViewSelect+" WHERE l.returned_at IS NULL AND l.due_at < ? ORDER BY l.due_at ASC, l.id ASC", now).Scan(&rows).Error.
	err := r.db.Raw(loanViewSelect+" WHERE l.returned_at IS NULL AND l.due_at < ? ORDER BY l.due_at ASC, l.id ASC", now).Scan(&rows).Error
	// 3. If err != nil, return nil, translateError(err). Otherwise return rows, nil.
	if err != nil {
		return nil, err
	}
	return rows, nil
}
