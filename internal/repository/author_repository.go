package repository

import (
	"gorm.io/gorm"
	"libratrack/internal/model"
)

type AuthorRepository interface {
	Create(author *model.Author) error
	FindByID(id uint) (*model.Author, error)
	List() ([]model.Author, error)
}

type gormAuthorRepository struct {
	db *gorm.DB
}

func NewAuthorRepository(db *gorm.DB) AuthorRepository {
	return &gormAuthorRepository{db: db}
}

func (r *gormAuthorRepository) Create(author *model.Author) error {
	if err := r.db.Create(author).Error; err != nil {
		return translateError(err)
	}
	return nil
}

func (r *gormAuthorRepository) FindByID(id uint) (*model.Author, error) {
	var a model.Author
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, translateError(err)
	}
	return &a, nil
}

func (r *gormAuthorRepository) List() ([]model.Author, error) {
	var authors []model.Author
	if err := r.db.Order("full_name ASC").Find(&authors).Error; err != nil {
		return nil, translateError(err)
	}
	if authors == nil {
		authors = []model.Author{}
	}
	return authors, nil
}
