package repository

import "libratrack/internal/model"

// List returns a paginated list of books, optionally filtered by category ID.
func (r *gormBookRepository) List(categoryID uint, page, limit int) ([]model.Book, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	// Preload replaces 1+2N queries with 3 queries (books, authors, categories) regardless of the number of books.
	q := r.db.Preload("Authors").Preload("Category").Order("id ASC").Offset((page - 1) * limit).Limit(limit)
	if categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	}
	var books []model.Book
	if err := q.Find(&books).Error; err != nil {
		return nil, translateError(err)
	}
	if books == nil {
		books = []model.Book{}
	}
	return books, nil
}
