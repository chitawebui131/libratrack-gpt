package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"libratrack/internal/model"
)

type BookRepository interface {
	Create(book *model.Book, authorIDs []uint) error
	FindByID(id uint) (*model.Book, error)
	List(categoryID uint, page, limit int) ([]model.Book, error)
	Update(book *model.Book, authorIDs []uint) error
	Delete(id uint) error
}

type gormBookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) BookRepository {
	return &gormBookRepository{db: db}
}

func uniqueIDs(ids []uint) []uint {
	seen := make(map[uint]bool)
	var result []uint
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

func loadAuthors(tx *gorm.DB, ids []uint) ([]model.Author, error) {
	ids = uniqueIDs(ids)
	var authors []model.Author
	if err := tx.Where("id IN ?", ids).Find(&authors).Error; err != nil {
		return nil, err
	}
	if len(authors) != len(ids) {
		return nil, ErrInvalidReference
	}
	return authors, nil
}

func (r *gormBookRepository) Create(book *model.Book, authorIDs []uint) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		authors, err := loadAuthors(tx, authorIDs)
		if err != nil {
			return err
		}
		book.Authors = nil
		book.Category = nil
		book.Copies = nil
		if err := tx.Omit(clause.Associations).Create(book).Error; err != nil {
			return err
		}
		return tx.Model(book).Association("Authors").Replace(authors)
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

func (r *gormBookRepository) FindByID(id uint) (*model.Book, error) {
	var b model.Book
	if err := r.db.Preload("Authors").Preload("Category").Preload("Copies").First(&b, id).Error; err != nil {
		return nil, translateError(err)
	}
	return &b, nil
}

func (r *gormBookRepository) List(categoryID uint, page, limit int) ([]model.Book, error) {
	var books []model.Book
	query := r.db.Preload("Authors").Preload("Category").Preload("Copies")
	if categoryID != 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if err := query.Offset((page - 1) * limit).Limit(limit).Find(&books).Error; err != nil {
		return nil, translateError(err)
	}
	return books, nil
}

func (r *gormBookRepository) Update(book *model.Book, authorIDs []uint) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		authors, err := loadAuthors(tx, authorIDs)
		if err != nil {
			return err
		}
		if err := tx.Model(book).Updates(map[string]interface{}{
			"Title":         book.Title,
			"ISBN":          book.ISBN,
			"CategoryID":    book.CategoryID,
			"Publisher":     book.Publisher,
			"PublishedYear": book.PublishedYear,
			"Description":   book.Description,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(book).Association("Authors").Replace(authors); err != nil {
			return err
		}
		return nil
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
	if err := r.db.Delete(&model.Book{}, id).Error; err != nil {
		return translateError(err)
	}
	return nil
}
