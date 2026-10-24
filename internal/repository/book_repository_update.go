package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

func (r *gormBookRepository) Update(book *model.Book, authorIDs []uint) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		authors, err := loadAuthors(tx, authorIDs)
		if err != nil {
			return err
		}
		res := tx.Model(&model.Book{}).Where("id = ?", book.ID).Updates(map[string]any{
			"title":          book.Title,
			"isbn":           book.ISBN,
			"category_id":    book.CategoryID,
			"publisher":      book.Publisher,
			"published_year": book.PublishedYear,
			"description":    book.Description,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		var current model.Book
		if err := tx.First(&current, book.ID).Error; err != nil {
			return err
		}
		return tx.Model(&current).Association("Authors").Replace(authors)
	})
	if err != nil {
		return translateError(err)
	}
	var fresh model.Book
	if err := r.db.Preload("Authors").Preload("Category").First(&fresh, book.ID).Error; err != nil {
		return translateError(err)
	}
	*book = fresh
	return nil
}

func (r *gormBookRepository) Delete(id uint) error {
	res := r.db.Delete(&model.Book{}, id)
	if res.Error != nil {
		return translateError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
