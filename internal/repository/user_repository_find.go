package repository

import (
	"libratrack/internal/model"
)

func (r *gormUserRepository) FindByEmail(email string) (*model.User, error) {
	// 1. var u model.User.
	var u model.User
	// 2. err := r.db.Where("lower(email) = lower(?)", email).First(&u).Error. If err != nil, return nil, translateError(err).
	if err := r.db.Where("lower(email) = lower(?)", email).First(&u).Error; err != nil {
		return nil, translateError(err)
	}
	// 3. Return &u, nil.
	return &u, nil
}

func (r *gormUserRepository) List(page, limit int) ([]model.User, error) {
	// 1. offset, size := pagingBounds(page, limit).
	offset, size := pagingBounds(page, limit)
	// 2. var users []model.User.
	var users []model.User
	// 3. err := r.db.Order("id").Offset(offset).Limit(size).Find(&users).Error. If err != nil, return nil, translateError(err).
	if err := r.db.Order("id").Offset(offset).Limit(size).Find(&users).Error; err != nil {
		return nil, translateError(err)
	}
	// 4. Return users, nil.
	return users, nil
}
