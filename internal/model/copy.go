package model

import (
	"gorm.io/gorm"
)

// Copy represents a physical copy of a book.
type Copy struct {
	gorm.Model
	BookID uint `gorm:"not null"`
	// Book is the associated book.
	// NOTE: Book and Book.Copies must never be preloaded together (JSON cycle).
	Book *Book `json:"book,omitempty" gorm:"foreignKey:BookID"`
}

// Book represents a title.
type Book struct {
	gorm.Model
	Title      string    `json:"title"`
	AuthorID   uint      `json:"author_id"`
	Author     *Author   `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	CategoryID uint      `json:"category_id"`
	Category   *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Copies     []Copy    `json:"copies"`
}

// Author represents the author of a book.
type Author struct {
	gorm.Model
	Name  string `json:"name"`
	Books []Book `json:"books"`
}

// Category represents a genre or category.
type Category struct {
	gorm.Model
	Name  string `json:"name"`
	Books []Book `json:"books"`
}
