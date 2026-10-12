package repository

import (
	"libratrack/internal/model"
)

// BookRepository defines the operations for managing books.
type BookRepository interface {
	Create(book *model.Book) error
	FindByID(id uint) (*model.Book, error)
	List(category string, page, limit int) ([]*model.Book, error)
	Update(book *model.Book) error
	Delete(id uint) error
}
