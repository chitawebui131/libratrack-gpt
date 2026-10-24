package service

import (
	"errors"
	"strings"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
	ErrAccountDisabled    = errors.New("account is disabled")
)

type AuthService struct {
	users  repository.UserRepository
	tokens *TokenManager
}

func NewAuthService(users repository.UserRepository, tokens *TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

func (s *AuthService) GetByID(id uint) (*model.User, error) {
	return s.users.FindByID(id)
}

// normalizeEmail trims spaces and lower-cases the email.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Logout clears any server-side session data if necessary.
// This satisfies the handler.WebAuthService interface requirement.
func (s *AuthService) Logout(token string) error {
	// If your design does not require server-side token revocation,
	// returning nil here is perfectly fine as middleware handles cookies.
	return nil
}

// Refresh always fails: refresh tokens are not issued, so there is nothing to rotate.
// The middleware then clears the cookies and treats the visitor as anonymous.
func (s *AuthService) Refresh(refreshToken string) (*model.AuthResponse, error) {
	return nil, ErrInvalidCredentials
}
