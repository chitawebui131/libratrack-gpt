package service

import (
	"errors"
	"strings"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

// Register creates a new user. The role is ALWAYS reader: it is never taken from the request.
func (s *AuthService) Register(req model.RegisterRequest) (*model.User, error) {
	// 1. hash, err := HashPassword(req.Password). If err != nil, return nil, err.
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 2. user := &model.User{Email: normalizeEmail(req.Email), FullName: strings.TrimSpace(req.FullName), PasswordHash: hash, Role: model.RoleReader}.
	user := &model.User{
		Email:        normalizeEmail(req.Email),
		FullName:     strings.TrimSpace(req.FullName),
		PasswordHash: hash,
		Role:         model.RoleReader,
	}

	// 3. err = s.users.Create(user).
	err = s.users.Create(user)

	// 4. If errors.Is(err, repository.ErrConflict), return nil, ErrEmailTaken.
	if errors.Is(err, repository.ErrConflict) {
		return nil, ErrEmailTaken
	}

	// 5. If err != nil, return nil, err.
	if err != nil {
		return nil, err
	}

	// 6. Return user, nil.
	return user, nil
}
