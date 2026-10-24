package service

import (
	"errors"
	"fmt"
	"libratrack/internal/model"
	"libratrack/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Login checks the credentials and returns a token.
// Unknown email and wrong password must return the SAME error.
func (s *AuthService) Login(email, password string) (*model.AuthResponse, error) {
	// 1. user, err := s.users.FindByEmail(normalizeEmail(email)).
	user, err := s.users.FindByEmail(normalizeEmail(email))
	// 2. If errors.Is(err, repository.ErrNotFound): call CheckPasswordDummy(password) and return nil, ErrInvalidCredentials.
	if errors.Is(err, repository.ErrNotFound) {
		CheckPasswordDummy(password)
		return nil, ErrInvalidCredentials
	}
	// 3. If err != nil, return nil, err.
	if err != nil {
		return nil, err
	}
	// 4. If !CheckPassword(user.PasswordHash, password), return nil, ErrInvalidCredentials.
	if !CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	// 5. If user.Disabled, return nil, ErrAccountDisabled.
	if user.Disabled {
		return nil, ErrAccountDisabled
	}
	// 6. token, _, err := s.tokens.Issue(user.ID, user.Role). If err != nil, return nil, err.
	claims := &Claims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokens.TTL())),
		},
	}

	// Call Issue,
	token, err := s.tokens.Issue(claims)
	if err != nil {
		return nil, err
	}
	// 7. Return &model.AuthResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: int(s.tokens.TTL().Seconds()), User: *user}, nil.
	return &model.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.tokens.TTL().Seconds()),
		User:        *user,
	}, nil
}
