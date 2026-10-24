package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

// PanelRepository holds the write operations of the librarian panel.
type PanelRepository interface {
	AddCopy(bookID uint, inventoryNo, department string) error
	UpdateCopyStatus(copyID uint, status string) error
	UpdateCopy(copyID uint, inventoryNo, department string) error
	DeleteCopy(copyID uint) error
	SaveBook(b *model.Book, authorIDs []uint) error
	DeleteBook(id uint) error
}

type gormPanelRepository struct {
	db *gorm.DB
}
