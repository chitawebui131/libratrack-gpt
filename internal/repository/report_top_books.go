package repository

import "libratrack/internal/model"

// TopBooks returns the most borrowed books, most popular first.
func (r *gormReportRepository) TopBooks(limit int) ([]model.TopBook, error) {
	var rows []model.TopBook
	err := r.db.Raw(`
		SELECT b.id AS book_id, b.title AS title, COUNT(l.id) AS loan_count
		FROM loans l
		JOIN copies c ON c.id = l.copy_id
		JOIN books b ON b.id = c.book_id
		GROUP BY b.id, b.title
		ORDER BY loan_count DESC, b.id ASC
		LIMIT ?`, limit).Scan(&rows).Error
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}
