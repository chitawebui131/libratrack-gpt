package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

// DictionaryRepository reads and writes authors and categories for the web UI.
type DictionaryRepository interface {
	ListAuthors() ([]model.Author, error)
	GetAuthor(id uint) (*model.Author, error)
	CreateAuthor(a *model.Author) error
	UpdateAuthor(a *model.Author) error
	DeleteAuthor(id uint) error
	ListCategories() ([]model.Category, error)
	GetCategory(id uint) (*model.Category, error)
	CreateCategory(c *model.Category) error
	UpdateCategory(c *model.Category) error
	DeleteCategory(id uint) error
}

type gormDictionaryRepository struct {
	db *gorm.DB
}
