package model

import "time"

// User exists only for the schema and foreign keys in theme 8; authentication is implemented in theme 9.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"size:255;not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"size:255;not null;default:''" json:"-"`
	FullName     string    `gorm:"size:255;not null;default:''" json:"full_name"`
	Role         string    `gorm:"size:20;not null;default:reader" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
