package repository

import (
	"libratrack/internal/model"

	"gorm.io/gorm"
)

// UserRepository defines methods for user persistence.
type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	List(page, limit int) ([]model.User, error)
	UpdateRole(id uint, role string) (*model.User, error)
	UpdateProfile(id uint, email, fullName string) error
	SetPasswordHash(id uint, hash string) error
	SetDisabled(id uint, disabled bool) error
	Delete(id uint) error
}

// gormUserRepository implements UserRepository using GORM.
type gormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
//func NewUserRepository(db *gorm.DB) UserRepository {
//	return &gormUserRepository{db: db}
//}

// Create inserts a new user into the database.
func (r *gormUserRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return translateError(err)
	}
	return nil
}

// FindByID retrieves a user by its ID.
func (r *gormUserRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, translateError(err)
	}
	return &u, nil
}
