package model

import "time"

type Loan struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	CopyID     uint       `gorm:"not null;index" json:"copy_id"`
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	BorrowedAt time.Time  `gorm:"not null" json:"borrowed_at"`
	DueAt      time.Time  `gorm:"not null" json:"due_at"`
	ReturnedAt *time.Time `json:"returned_at"`
	// FineAmount is whole UAH. Money is an integer, never a float.
	FineAmount int   `gorm:"not null;default:0" json:"fine_amount"`
	Copy       *Copy `gorm:"foreignKey:CopyID" json:"copy,omitempty"`
	User       *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Loan) TableName() string { return "loans" }
