package model

// User represents a user in the system, used for authentication and Role-Based Access Control (RBAC).
type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	FullName     string `json:"full_name"`
	Email        string `gorm:"unique" json:"email"`
	PasswordHash string `json:"-"` // Never serialize password hash
	Role         string `json:"role"`
}

// Role constants
const (
	RoleAdmin     = "admin"
	RoleLibrarian = "librarian"
	RoleReader    = "reader"
)

// IsValidRole checks if the provided role string is one of the defined roles.
func IsValidRole(role string) bool {
	return role == RoleAdmin || role == RoleLibrarian || role == RoleReader
}

// RegisterRequest holds data required for user registration.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"` // 72 is the bcrypt input limit in bytes
	FullName string `json:"full_name" binding:"required,max=255"`
}

// LoginRequest holds data required for user login.
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

// UpdateUserRoleRequest holds data required to update a user's role.
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin librarian reader"`
}
