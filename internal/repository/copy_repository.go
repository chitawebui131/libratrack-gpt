package repository

import (
	"context"
	"libratrack/internal/model"
	"time"

	"gorm.io/gorm"
)

// CopyRepository defines the interface for copy data access operations.
type CopyRepository interface {
	GetCopyByID(ctx context.Context, id uint) (*model.Copy, error)
	GetCopiesByBookID(ctx context.Context, bookID uint) ([]model.Copy, error)
	UpdateCopyStatus(ctx context.Context, copyID uint, status model.CopyStatus) error
	// Other methods...
}

type copyRepository struct {
	DB *gorm.DB
}

// NewCopyRepository creates a new CopyRepository instance.
func NewCopyRepository(db *gorm.DB) CopyRepository {
	return &copyRepository{DB: db}
}

// GetCopyByID retrieves a copy by its ID.
func (r *copyRepository) GetCopyByID(ctx context.Context, id uint) (*model.Copy, error) {
	var copy model.Copy
	err := r.DB.WithContext(ctx).First(&copy, id).Error
	if err != nil {
		return nil, err
	}
	return &copy, nil
}

// GetCopiesByBookID retrieves all copies associated with a specific book ID.
func (r *copyRepository) GetCopiesByBookID(ctx context.Context, bookID uint) ([]model.Copy, error) {
	var copies []model.Copy
	err := r.DB.WithContext(ctx).Where("book_id = ?", bookID).Find(&copies).Error
	if err != nil {
		return nil, err
	}
	return copies, nil
}

// UpdateCopyStatus updates the status of a copy.
func (r *copyRepository) UpdateCopyStatus(ctx context.Context, copyID uint, status model.CopyStatus) error {
	// Note: In a real application, we would validate the status against allowed transitions.
	result := r.DB.WithContext(ctx).Model(&model.Copy{}).Where("id = ?", copyID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
