package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

type CopyRepository interface {
	Create(copy *model.Copy) error
	ListByBook(bookID uint) ([]model.Copy, error)
	Availability(bookID uint) (*model.BookAvailability, error)
}

type gormCopyRepository struct {
	db *gorm.DB
}

func NewCopyRepository(db *gorm.DB) CopyRepository {
	return &gormCopyRepository{db: db}
}

func (r *gormCopyRepository) ensureBookExists(bookID uint) error {
	var n int64
	if err := r.db.Model(&model.Book{}).Where("id = ?", bookID).Count(&n).Error; err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *gormCopyRepository) Create(copy *model.Copy) error {
	if err := r.ensureBookExists(copy.BookID); err != nil {
		return err
	}
	if copy.Status == "" {
		copy.Status = model.CopyStatusAvailable
	}
	if err := r.db.Create(copy).Error; err != nil {
		return translateError(err)
	}
	return nil
}

func (r *gormCopyRepository) ListByBook(bookID uint) ([]model.Copy, error) {
	if err := r.ensureBookExists(bookID); err != nil {
		return nil, err
	}
	var copies []model.Copy
	if err := r.db.Where("book_id = ?", bookID).Order("id ASC").Find(&copies).Error; err != nil {
		return nil, translateError(err)
	}
	if copies == nil {
		copies = []model.Copy{}
	}
	return copies, nil
}

func (r *gormCopyRepository) Availability(bookID uint) (*model.BookAvailability, error) {
	if err := r.ensureBookExists(bookID); err != nil {
		return nil, err
	}
	result := model.BookAvailability{BookID: bookID}
	if err := r.db.Raw("SELECT COUNT(*) AS total_copies, COUNT(*) FILTER (WHERE status = 'available') AS available_copies FROM copies WHERE book_id = ?", bookID).Scan(&result).Error; err != nil {
		return nil, translateError(err)
	}
	return &result, nil
}
