package repository

import (
	"context"
	"libratrack/internal/model"

	"gorm.io/gorm"
)

// CopyRepository defines the interface for copy data access operations.
type CopyRepository interface {
	GetCopyByID(ctx context.Context, id uint) (*model.Copy, error)
	GetCopiesByBookID(ctx context.Context, bookID uint) ([]model.Copy, error)
	UpdateCopyStatus(ctx context.Context, copyID uint, status string) error

	// Methods used by the copy handlers:
	Create(ctx context.Context, copy *model.Copy) error
	ListByBook(ctx context.Context, bookID uint) ([]model.Copy, error)
	Availability(ctx context.Context, bookID uint) (*model.BookAvailability, error)
}

type copyRepository struct {
	DB *gorm.DB
}

// NewCopyRepository creates a new CopyRepository instance.
func NewCopyRepository(db *gorm.DB) CopyRepository {
	return &copyRepository{DB: db}
}

func (r *copyRepository) GetCopyByID(ctx context.Context, id uint) (*model.Copy, error) {
	var cp model.Copy
	err := r.DB.WithContext(ctx).First(&cp, id).Error
	if err != nil {
		return nil, translateError(err)
	}
	return &cp, nil
}

func (r *copyRepository) GetCopiesByBookID(ctx context.Context, bookID uint) ([]model.Copy, error) {
	var copies []model.Copy
	err := r.DB.WithContext(ctx).Where("book_id = ?", bookID).Find(&copies).Error
	if err != nil {
		return nil, translateError(err)
	}
	return copies, nil
}

func (r *copyRepository) UpdateCopyStatus(ctx context.Context, copyID uint, status string) error {
	result := r.DB.WithContext(ctx).Model(&model.Copy{}).Where("id = ?", copyID).Update("status", status)
	if result.Error != nil {
		return translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *copyRepository) Create(ctx context.Context, cp *model.Copy) error {
	return translateError(r.DB.WithContext(ctx).Omit("Book").Create(cp).Error)
}

func (r *copyRepository) ListByBook(ctx context.Context, bookID uint) ([]model.Copy, error) {
	var copies []model.Copy
	err := r.DB.WithContext(ctx).Where("book_id = ?", bookID).Find(&copies).Error
	if err != nil {
		return nil, translateError(err)
	}
	if copies == nil {
		copies = []model.Copy{}
	}
	return copies, nil
}

func (r *copyRepository) Availability(ctx context.Context, bookID uint) (*model.BookAvailability, error) {
	var total int64
	var available int64

	err := r.DB.WithContext(ctx).Model(&model.Copy{}).Where("book_id = ?", bookID).Count(&total).Error
	if err != nil {
		return nil, translateError(err)
	}

	err = r.DB.WithContext(ctx).Model(&model.Copy{}).Where("book_id = ? AND status = ?", bookID, model.CopyStatusAvailable).Count(&available).Error
	if err != nil {
		return nil, translateError(err)
	}

	return &model.BookAvailability{
		TotalCopies:     int(total),
		AvailableCopies: int(available),
	}, nil
}
