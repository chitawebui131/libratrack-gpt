package service

import (
	"errors"
	"strings"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

// CreateUser creates a user with any role (admin panel only).
func (s *AuthService) CreateUser(email, password, fullName, role string) (*model.User, error) {
	// 1. If !model.IsValidRole(role) return nil, errors.New("invalid role").
	if !model.IsValidRole(role) {
		return nil, errors.New("invalid role")
	}
	// 2. If len(password) < 8 or len(password) > 72 return nil, ErrPasswordInvalid.
	if len(password) < 8 || len(password) > 72 {
		return nil, ErrPasswordInvalid
	}
	// 3. hash, err := HashPassword(password). Return the error if any.
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	// 4. user := &model.User{Email: normalizeEmail(email), FullName: strings.TrimSpace(fullName), PasswordHash: hash, Role: role}.
	user := &model.User{
		Email:        normalizeEmail(email),
		FullName:     strings.TrimSpace(fullName),
		PasswordHash: hash,
		Role:         role,
	}
	// 5. err = s.users.Create(user). If errors.Is(err, repository.ErrConflict) return nil, ErrEmailTaken. Other error: return nil, err.
	err = s.users.Create(user)
	if errors.Is(err, repository.ErrConflict) {
		return nil, ErrEmailTaken
	}
	if err != nil {
		return nil, err
	}
	// 6. Return user, nil.
	return user, nil
}

// UpdateProfile saves the email (normalized) and the full name. A duplicate email gives ErrEmailTaken.
func (s *AuthService) UpdateProfile(id uint, email, fullName string) error {
	// err := s.users.UpdateProfile(id, normalizeEmail(email), strings.TrimSpace(fullName)). ErrConflict -> ErrEmailTaken. Otherwise return err.
	err := s.users.UpdateProfile(id, normalizeEmail(email), strings.TrimSpace(fullName))
	if errors.Is(err, repository.ErrConflict) {
		return ErrEmailTaken
	}
	return err
}

// ResetPassword sets a new password (8 to 72 bytes).
func (s *AuthService) ResetPassword(id uint, password string) error {
	// 1. If len(password) < 8 or len(password) > 72 return ErrPasswordInvalid.
	if len(password) < 8 || len(password) > 72 {
		return ErrPasswordInvalid
	}
	// 2. hash, err := HashPassword(password). Return the error if any.
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	// 3. Return s.users.SetPasswordHash(id, hash).
	return s.users.SetPasswordHash(id, hash)
}
