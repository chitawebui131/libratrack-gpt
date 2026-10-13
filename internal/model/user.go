package model

import (
	"time"
)

// User represents a user in the system, used for authentication and Role-Based Access Control (RBAC).
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"unique;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"` // PasswordHash must keep the JSON tag "-" so it is never serialized.
	FullName     string    `gorm:"not null" json:"full_name"`
	Role         string    `gorm:"default:reader" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Role constants
const (
	RoleAdmin    = "admin"
	RoleLibrarian = "librarian"
	RoleReader   = "reader"
)

// IsValidRole checks if the provided role string is one of the defined roles.
func IsValidRole(role string) bool {
	return role == RoleAdmin || role == RoleLibrarian || role == RoleReader
}

// RegisterRequest defines the input structure for user registration.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"` // 72 is the bcrypt input limit in bytes.
	FullName string `json:"full_name" binding:"required,max=255"`
}

// LoginRequest defines the input structure for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse defines the structure returned upon successful authentication.
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	User        User   `json:"user"`
}

// UpdateUserRoleRequest defines the input structure for updating a user's role.
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin librarian reader"`
}
