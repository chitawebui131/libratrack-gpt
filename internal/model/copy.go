package model

import (
	"time"
)

// Copy represents a physical copy of a book.
type Copy struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	BookID      uint      `gorm:"not null" json:"book_id"`
	Status      string    `gorm:"size:20;not null" json:"status"`
	InventoryNo string    `gorm:"size:255;uniqueIndex;not null" json:"inventory_no"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Department is the library department or storage where this copy is kept.
	Department string `gorm:"column:department" json:"department,omitempty"`

	// Book and Book.Copies must never be preloaded together (JSON cycle).
	Book *Book `json:"book,omitempty" gorm:"foreignKey:BookID"`
}

const (
	CopyStatusAvailable   = "available"
	CopyStatusLost        = "lost"
	CopyStatusUnderRepair = "under_repair"
)
