package model

import "time"

type CopyStatus string

const (
    CopyStatusAvailable   CopyStatus = "available"
    CopyStatusBorrowed    CopyStatus = "borrowed"
    CopyStatusLost        CopyStatus = "lost"
    CopyStatusUnderRepair CopyStatus = "under_repair"
)

type Copy struct {
    ID          uint       `gorm:"primaryKey" json:"id"`
    BookID      uint       `gorm:"not null;index" json:"book_id"`
    InventoryNo string     `gorm:"size:50;not null;uniqueIndex" json:"inventory_no"`
    Status      CopyStatus `gorm:"size:20;not null;default:available" json:"status"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateCopyRequest struct {
    InventoryNo string     `json:"inventory_no" binding:"required,max=50"`
    Status      CopyStatus `json:"status" binding:"omitempty,oneof=available borrowed lost under_repair"`
}

type BookAvailability struct {
    BookID          uint `json:"book_id"`
    TotalCopies     int  `json:"total_copies"`
    AvailableCopies int  `json:"available_copies"`
}
