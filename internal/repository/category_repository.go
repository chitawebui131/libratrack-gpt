package repository

import (
	"libratrack/internal/model"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *model.Category) error
	FindByID(id uint) (*model.Category, error)
	List() ([]model.Category, error)
}

type gormCategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &gormCategoryRepository{db: db}
}

func (r *gormCategoryRepository) Create(category *model.Category) error {
	if err := r.db.Create(category).Error; err != nil {
		return translateError(err)
	}
	return nil
}

func (r *gormCategoryRepository) FindByID(id uint) (*model.Category, error) {
	var c model.Category
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, translateError(err)
	}
	return &c, nil
}

func (r *gormCategoryRepository) List() ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.Order("name ASC").Find(&categories).Error; err != nil {
		return nil, translateError(err)
	}
	if categories == nil {
		categories = []model.Category{}
	}
	return categories, nil
}
