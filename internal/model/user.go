package model

import (
	"time"
)

// User is now used for authentication and RBAC.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	FullName     string    `gorm:"size:255;not null" json:"full_name"`
	Role         string    `gorm:"size:20;not null" json:"role"`
	Disabled     bool      `gorm:"column:disabled;not null;default:false" json:"disabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRegisterRequest is used to register a new user.
// The password field is limited to 72 bytes, which is the maximum input size for bcrypt.
type UserRegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	FullName string `json:"full_name" binding:"required,max=255"`
}

// UserLoginRequest is used to authenticate a user.
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserAuthResponse is returned after successful authentication.
type UserAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // seconds
	User        User   `json:"user"`
}

// UserUpdateUserRoleRequest is used to change a user's role.
type UserUpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin librarian reader"`
}
