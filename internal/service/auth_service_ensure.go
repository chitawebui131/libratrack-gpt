package service

import (
	"errors"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

// EnsureUser creates a user if it does not exist yet. It never changes an existing user.
func (s *AuthService) EnsureUser(email, password, fullName, role string) error {
	// 1. If !model.IsValidRole(role), return errors.New("invalid role").
	if !model.IsValidRole(role) {
		return errors.New("invalid role")
	}
	// 2. If len(password) < 8, return errors.New("password must be at least 8 characters").
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	// 3. email = normalizeEmail(email).
	email = normalizeEmail(email)
	// 4. _, err := s.users.FindByEmail(email). If err == nil, return nil (the user exists; change nothing).
	_, err := s.users.FindByEmail(email)
	if err == nil {
		return nil // the user exists; change nothing
	}
	// 5. If !errors.Is(err, repository.ErrNotFound), return err.
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	// 6. hash, err := HashPassword(password). If err != nil, return err.
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	// 7. Return s.users.Create(&model.User{Email: email, FullName: fullName, PasswordHash: hash, Role: role}).
	return s.users.Create(&model.User{Email: email, FullName: fullName, PasswordHash: hash, Role: role})
}
