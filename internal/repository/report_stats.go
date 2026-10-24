package repository

import (
	"time"

	"libratrack/internal/model"
)

// Stats returns simple counters for the dashboard.
func (r *gormReportRepository) Stats(now time.Time) (*model.DashboardStats, error) {
	// 1. s := &model.DashboardStats{}.
	s := &model.DashboardStats{}
	// 2. Run these five queries with r.db.Raw(sql, args...).Scan(&target).Error. On any error return nil, translateError(err):
	//      "SELECT COUNT(*) FROM books"                                              -> s.TotalBooks
	//      "SELECT COUNT(*) FROM copies"                                             -> s.TotalCopies
	//      "SELECT COUNT(*) FROM copies WHERE status = 'available'"                  -> s.AvailableCopies
	//      "SELECT COUNT(*) FROM loans WHERE returned_at IS NULL"                    -> s.ActiveLoans
	//      "SELECT COUNT(*) FROM loans WHERE returned_at IS NULL AND due_at < ?" (arg now) -> s.OverdueLoans
	err := r.db.Raw("SELECT COUNT(*) FROM books").Scan(&s.TotalBooks).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Raw("SELECT COUNT(*) FROM copies").Scan(&s.TotalCopies).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Raw("SELECT COUNT(*) FROM copies WHERE status = 'available'").Scan(&s.AvailableCopies).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Raw("SELECT COUNT(*) FROM loans WHERE returned_at IS NULL").Scan(&s.ActiveLoans).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Raw("SELECT COUNT(*) FROM loans WHERE returned_at IS NULL AND due_at < ?", now).Scan(&s.OverdueLoans).Error
	if err != nil {
		return nil, err
	}
	// 3. Return s, nil.
	return s, nil
}
