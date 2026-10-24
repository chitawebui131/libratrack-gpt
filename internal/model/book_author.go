package model

type BookAuthor struct {
	BookID   uint `gorm:"primaryKey"`
	AuthorID uint `gorm:"primaryKey"`
}
