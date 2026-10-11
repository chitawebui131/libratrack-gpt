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
