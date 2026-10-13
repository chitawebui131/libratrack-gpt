package model

import "time"

// CopyStatus defines the possible states of a copy.
type CopyStatus string

const (
	CopyStatusAvailable CopyStatus = "available"
	CopyStatusLoaned    CopyStatus = "loaned"
	CopyStatusDamaged   CopyStatus = "damaged"
	CopyStatusLost      CopyStatus = "lost"
)

// Copy represents a physical copy of a book.
type Copy struct {
	ID     uint       `gorm:"primaryKey" json:"id"`
	BookID uint       `gorm:"not null" json:"book_id"`
	Status CopyStatus `gorm:"default:available" json:"status"` // Status field added to resolve repository errors.
	// Book field added to link Copy to Book.
	Book      *Book     `json:"book,omitempty" gorm:"foreignKey:BookID"` // Book and Book.Copies must never be preloaded together (JSON cycle).
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Book represents a title.
type Book struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"not null" json:"title"`
	AuthorID  uint      `gorm:"not null" json:"author_id"`
	Author    Author    `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	Copies    []Copy    `json:"copies,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Author represents the author of a book.
type Author struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Books     []Book    `json:"books,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
