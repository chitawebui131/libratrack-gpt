package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type fakeUserRepo struct {
	repository.UserRepository
	byEmail   map[string]*model.User
	createErr error
}

func (f *fakeUserRepo) Create(u *model.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	// Verify that email constraints are evaluated correctly in test profiles
	if _, exists := f.byEmail[u.Email]; exists {
		return repository.ErrConflict
	}
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeUserRepo) FindByEmail(email string) (*model.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func TestCreateUserValidates(t *testing.T) {
	originalBcryptCost := bcryptCost
	t.Cleanup(func() {
		bcryptCost = originalBcryptCost
	})
	bcryptCost = bcrypt.MinCost

	fake := &fakeUserRepo{
		byEmail: make(map[string]*model.User),
	}

	tokens, err := NewTokenManager(strings.Repeat("k", 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	authService := NewAuthService(fake, tokens)

	// Test password validation checks first (Matches new code structure)
	_, err = authService.CreateUser("test@example.com", "pass123", "Test User", "user")
	if !errors.Is(err, ErrPasswordInvalid) {
		t.Errorf("Expected ErrPasswordInvalid for short password, got %v", err)
	}

	_, err = authService.CreateUser("test@example.com", strings.Repeat("a", 73), "Test User", "user")
	if !errors.Is(err, ErrPasswordInvalid) {
		t.Errorf("Expected ErrPasswordInvalid for long password, got %v", err)
	}

	// Test invalid role parameter boundary processing
	_, err = authService.CreateUser("test@example.com", "password123", "Test User", "invalid_role")
	if err == nil {
		t.Error("Expected error for invalid role")
	}

	// Test valid user creation execution path
	user, err := authService.CreateUser("Test@Example.Com", "password123", "  Test User  ", "admin")
	if err != nil {
		t.Fatal(err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email to be lowercased and trimmed, got %s", user.Email)
	}

	if user.FullName != "Test User" {
		t.Errorf("Expected full name to be trimmed, got %s", user.FullName)
	}

	if user.Role != "admin" {
		t.Errorf("Expected role to be admin, got %s", user.Role)
	}

	if user.PasswordHash == "" || user.PasswordHash == "password123" {
		t.Error("Password hashing security checks failed")
	}
}

func TestCreateUserEmailTaken(t *testing.T) {

	originalBcryptCost := bcryptCost
	t.Cleanup(func() {
		bcryptCost = originalBcryptCost
	})
	bcryptCost = bcrypt.MinCost


	fake := &fakeUserRepo{
		byEmail:   make(map[string]*model.User),
		createErr: repository.ErrConflict,
	}


	tokens, err := NewTokenManager(strings.Repeat("k", 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	authService := NewAuthService(fake, tokens)


	_, err = authService.CreateUser("test@example.com", "password123", "Test User", "admin")
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("Expected ErrEmailTaken when email is taken, got %v", err)
	}
}


func TestLoginDisabledUser(t *testing.T) {
	originalBcryptCost := bcryptCost
	t.Cleanup(func() {
		bcryptCost = originalBcryptCost
	})
	bcryptCost = bcrypt.MinCost

	fake := &fakeUserRepo{
		byEmail: make(map[string]*model.User),
	}

	tokens, err := NewTokenManager(strings.Repeat("k", 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	authService := NewAuthService(fake, tokens)

	hashedPassword, _ := HashPassword("password1")
	disabledUser := &model.User{
		ID:           1,
		Email:        "test@example.com",
		FullName:     "Test User",
		PasswordHash: hashedPassword,
		Role:         "user",
		Disabled:     true,
	}
	fake.byEmail["test@example.com"] = disabledUser

	_, err = authService.Login("test@example.com", "password1")
	if !errors.Is(err, ErrAccountDisabled) {
		t.Errorf("Expected ErrAccountDisabled for disabled user, got %v", err)
	}

	_, err = authService.Login("test@example.com", "wrongpassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials for wrong password, got %v", err)
	}
}
