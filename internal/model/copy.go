package model

// Copy represents a physical copy of a book.
type Copy struct {
	ID     uint `gorm:"primaryKey"`
	BookID uint `gorm:"index;not null"`
	// ... other fields
	Book *Book `json:"book,omitempty" gorm:"foreignKey:BookID"` // Book and Book.Copies must never be preloaded together (JSON cycle)
}
