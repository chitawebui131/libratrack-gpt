package repository

import (
	"time"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

type RefreshTokenRepository interface {
	RevokeAllForUser(userID uint) error
}

type gormRefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &gormRefreshTokenRepository{db: db}
}

func (r *gormRefreshTokenRepository) RevokeAllForUser(userID uint) error {
	err := r.db.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now()).Error
	return translateError(err)
}
