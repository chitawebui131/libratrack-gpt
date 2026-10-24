package repository

import (
	"libratrack/internal/model"
)

// ListViewsByUser returns all loans of one user (active first, then newest first) and the total count.
func (r *gormLoanViewRepository) ListViewsByUser(userID uint, page, limit int) ([]model.LoanView, int64, error) {
	// 1. offset, size := pagingBounds(page, limit).
	offset, size := pagingBounds(page, limit)
	// 2. var total int64. err := r.db.Raw("SELECT COUNT(*) FROM loans WHERE user_id = ?", userID).Scan(&total).Error. On error return nil, 0, translateError(err).
	var total int64
	err := r.db.Raw("SELECT COUNT(*) FROM loans WHERE user_id = ?", userID).Scan(&total).Error
	if err != nil {
		return nil, 0, translateError(err)
	}
	// 3. var rows []model.LoanView.
	var rows []model.LoanView
	// 4. err = r.db.Raw(loanViewSelect+" WHERE l.user_id = ? ORDER BY (l.returned_at IS NULL) DESC, l.borrowed_at DESC, l.id DESC LIMIT ? OFFSET ?", userID, size, offset).Scan(&rows).Error. On error return nil, 0, translateError(err).
	err = r.db.Raw(loanViewSelect+" WHERE l.user_id = ? ORDER BY (l.returned_at IS NULL) DESC, l.borrowed_at DESC, l.id DESC LIMIT ? OFFSET ?", userID, size, offset).Scan(&rows).Error
	if err != nil {
		return nil, 0, translateError(err)
	}
	// 5. Return rows, total, nil.
	return rows, total, nil
}

// ListActiveViews returns loans with returned_at IS NULL ordered by due date, and the total count.
func (r *gormLoanViewRepository) ListActiveViews(page, limit int) ([]model.LoanView, int64, error) {
	// 1. offset, size := pagingBounds(page, limit).
	offset, size := pagingBounds(page, limit)
	// 2. var total int64. err := r.db.Raw("SELECT COUNT(*) FROM loans WHERE returned_at IS NULL").Scan(&total).Error. On error return nil, 0, translateError(err).
	var total int64
	err := r.db.Raw("SELECT COUNT(*) FROM loans WHERE returned_at IS NULL").Scan(&total).Error
	if err != nil {
		return nil, 0, translateError(err)
	}
	// 3. var rows []model.LoanView.
	var rows []model.LoanView
	// 4. err = r.db.Raw(loanViewSelect+" WHERE l.returned_at IS NULL ORDER BY l.due_at ASC, l.id ASC LIMIT ? OFFSET ?", size, offset).Scan(&rows).Error. On error return nil, 0, translateError(err).
	err = r.db.Raw(loanViewSelect+" WHERE l.returned_at IS NULL ORDER BY l.due_at ASC, l.id ASC LIMIT ? OFFSET ?", size, offset).Scan(&rows).Error
	if err != nil {
		return nil, 0, translateError(err)
	}
	// 5. Return rows, total, nil.
	return rows, total, nil
}
