package repository

import "libratrack/internal/model"

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
	q := r.db.Order("id ASC").Offset((page - 1) * limit).Limit(limit)
	if categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	}
	var books []model.Book
	if err := q.Find(&books).Error; err != nil {
		return nil, translateError(err)
	}
	// NAIVE: N+1 queries, fixed in step 8.23
	for i := range books {
		if err := r.db.Model(&books[i]).Association("Authors").Find(&books[i].Authors); err != nil {
			return nil, translateError(err)
		}
		var cat model.Category
		if err := r.db.First(&cat, books[i].CategoryID).Error; err == nil {
			books[i].Category = &cat
		}
	}
	if books == nil {
		books = []model.Book{}
	}
	return books, nil
}
