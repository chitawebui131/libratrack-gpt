package model

import "time"

// Book represents a library catalog entry.
// Author and Category are temporary plain-text fields and will become relational in theme 8.
type Book struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	ISBN          string    `json:"isbn"`
	Author        string    `json:"author"`
	Category      string    `json:"category"`
	PublishedYear int       `json:"published_year,omitempty"`
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateBookRequest represents the payload for creating a new book.
// All fields except PublishedYear and Description are required.
type CreateBookRequest struct {
	Title         string `json:"title" binding:"required"`
	ISBN          string `json:"isbn" binding:"required"`
	Author        string `json:"author" binding:"required"`
	Category      string `json:"category" binding:"required"`
	PublishedYear int    `json:"published_year"`
	Description   string `json:"description"`
}

// UpdateBookRequest represents the payload for updating an existing book.
// All fields except PublishedYear and Description are required.
type UpdateBookRequest struct {
	Title         string `json:"title" binding:"required"`
	ISBN          string `json:"isbn" binding:"required"`
	Author        string `json:"author" binding:"required"`
	Category      string `json:"category" binding:"required"`
	PublishedYear int    `json:"published_year"`
	Description   string `json:"description"`
}
