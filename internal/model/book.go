package model

import "time"

type Book struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Title         string    `gorm:"size:255;not null" json:"title"`
	ISBN          string    `gorm:"size:20;not null;uniqueIndex" json:"isbn"`
	CategoryID    uint      `gorm:"not null;index" json:"category_id"`
	Category      *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Publisher     string    `gorm:"size:255;not null;default:''" json:"publisher"`
	PublishedYear int       `gorm:"not null;default:0" json:"published_year"`
	Description   string    `gorm:"type:text;not null;default:''" json:"description"`
	Authors       []Author  `gorm:"many2many:book_authors" json:"authors,omitempty"`
	Copies        []Copy    `gorm:"foreignKey:BookID" json:"copies,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateBookRequest struct {
	Title         string `json:"title" binding:"required,max=255"`
	ISBN          string `json:"isbn" binding:"required,min=10,max=20"`
	CategoryID    uint   `json:"category_id" binding:"required,gt=0"`
	AuthorIDs     []uint `json:"author_ids" binding:"required,min=1,dive,gt=0"`
	Publisher     string `json:"publisher" binding:"max=255"`
	PublishedYear int    `json:"published_year" binding:"gte=0,lte=2100"`
	Description   string `json:"description" binding:"max=5000"`
}

type UpdateBookRequest struct {
	Title         string `json:"title" binding:"required,max=255"`
	ISBN          string `json:"isbn" binding:"required,min=10,max=20"`
	CategoryID    uint   `json:"category_id" binding:"required,gt=0"`
	AuthorIDs     []uint `json:"author_ids" binding:"required,min=1,dive,gt=0"`
	Publisher     string `json:"publisher" binding:"max=255"`
	PublishedYear int    `json:"published_year" binding:"gte=0,lte=2100"`
	Description   string `json:"description" binding:"max=5000"`
}
