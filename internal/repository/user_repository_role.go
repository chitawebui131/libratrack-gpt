package repository

import (
	"libratrack/internal/model"

	"gorm.io/gorm"
)

// UpdateRole updates a user's role and returns the updated user.
func (r *gormUserRepository) UpdateRole(id uint, role string) (*model.User, error) {
	// 1. res := r.db.Model(&model.User{}).Where("id = ?", id).Update("role", role).
	res := r.db.Model(&model.User{}).Where("id = ?", id).Update("role", role)
	// 2. If res.Error != nil, return nil, translateError(res.Error).
	if res.Error != nil {
		return nil, translateError(res.Error)
	}
	// 3. If res.RowsAffected == 0, return nil, ErrNotFound.
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	// 4. Return r.FindByID(id).
	return r.FindByID(id)
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	// Return &gormUserRepository{db: db}.
	return &gormUserRepository{db: db}
}
