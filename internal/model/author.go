package model

import "time"

type Author struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FullName  string    `gorm:"size:255;not null" json:"full_name"`
	Bio       string    `gorm:"type:text;not null;default:''" json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAuthorRequest struct {
	FullName string `json:"full_name" binding:"required,max=255"`
	Bio      string `json:"bio" binding:"max=2000"`
}
