package repository

import (
	"libratrack/internal/model"
)

func (r *gormCatalogRepository) Search(f BookFilter) ([]model.Book, int64, error) {
	// 1. offset, size := pagingBounds(f.Page, f.Limit).
	// 2. var total int64. err := applyBookFilter(r.db.Model(&model.Book{}), f).Count(&total).Error. If err != nil, return nil, 0, translateError(err).
	// 3. var books []model.Book.
	// 4. err = applyBookFilter(r.db.Model(&model.Book{}), f).Preload("Authors").Preload("Category").Order("books.title ASC, books.id ASC").Offset(offset).Limit(size).Find(&books).Error.
	//    If err != nil, return nil, 0, translateError(err).
	// 5. Return books, total, nil.
	offset, size := pagingBounds(f.Page, f.Limit)
	var total int64
	err := applyBookFilter(r.db.Model(&model.Book{}), f).Count(&total).Error
	if err != nil {
		return nil, 0, translateError(err)
	}
	var books []model.Book
	err = applyBookFilter(r.db.Model(&model.Book{}), f).Preload("Authors").Preload("Category").Order("books.title ASC, books.id ASC").Offset(offset).Limit(size).Find(&books).Error
	if err != nil {
		return nil, 0, translateError(err)
	}
	return books, total, nil
}
