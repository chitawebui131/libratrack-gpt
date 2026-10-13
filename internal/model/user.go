package model

// User represents a user in the system, used for authentication and Role-Based Access Control (RBAC).
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"column:password_hash"` // JSON tag "-" ensures it is never serialized
	FullName     string `gorm:"not null"`
	Role         string `gorm:"type:varchar(50);not null"`
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

// RegisterRequest is used for user registration.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"` // 72 is the bcrypt input limit in bytes
	FullName string `json:"full_name" binding:"required,max=255"`
}

// LoginRequest is used for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is the structure returned upon successful authentication.
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	User        User   `json:"user"`
}

// UpdateUserRoleRequest is used to change a user's role.
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin librarian reader"`
}
