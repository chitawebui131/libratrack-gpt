package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"libratrack/internal/model"
)

// BookRepository defines the interface for book persistence operations.
type BookRepository interface {
	Create(book *model.Book, authorIDs []uint) error
	FindByID(id uint) (*model.Book, error)
	List(categoryID uint, page, limit int) ([]model.Book, error)
	Update(book *model.Book, authorIDs []uint) error
	Delete(id uint) error
	GetCategoryByName(name string) (*model.Category, error)
}

type gormBookRepository struct {
	db *gorm.DB
}

// NewBookRepository creates a new BookRepository backed by GORM.
func NewBookRepository(db *gorm.DB) BookRepository {
	return &gormBookRepository{db: db}
}

// uniqueIDs returns the input slice with duplicates removed, preserving order.
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

// loadAuthors loads authors by IDs, ensuring all referenced authors exist.
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

// Create inserts a new book and associates it with the provided authors.
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

// FindByID retrieves a book by its ID, including related entities.
func (r *gormBookRepository) FindByID(id uint) (*model.Book, error) {
	var b model.Book
	if err := r.db.Preload("Authors").Preload("Category").Preload("Copies").First(&b, id).Error; err != nil {
		return nil, translateError(err)
	}
	return &b, nil
}

// Update modifies an existing book and its author associations.
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

// Delete removes a book by its ID.
func (r *gormBookRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Book{}, id).Error; err != nil {
		return translateError(err)
	}
	return nil
}

// GetCategoryByName retrieves a category by its name.
func (r *gormBookRepository) GetCategoryByName(name string) (*model.Category, error) {
	var cat model.Category
	if err := r.db.Where("name = ?", name).First(&cat).Error; err != nil {
		return nil, translateError(err)
	}
	return &cat, nil
}
