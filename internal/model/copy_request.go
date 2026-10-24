package model

// CreateCopyRequest defines the structure for creating a new copy of a book.
type CreateCopyRequest struct {
	InventoryNo string `json:"inventory_no" binding:"required"`
	Status      string `json:"status" binding:"required"`
}
