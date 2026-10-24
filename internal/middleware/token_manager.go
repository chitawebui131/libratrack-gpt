package middleware

// TokenManager defines the interface for generating and validating JWT tokens.
type TokenManager interface {
	GenerateToken(userID uint) (string, error)
	ValidateToken(token string) (uint, error)
}
