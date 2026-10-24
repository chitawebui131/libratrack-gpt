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
	// Save original bcryptCost and restore it after the test
	originalBcryptCost := bcryptCost
	t.Cleanup(func() {
		bcryptCost = originalBcryptCost
	})
	bcryptCost = bcrypt.MinCost

	// Create a fake repository
	fake := &fakeUserRepo{
		byEmail: make(map[string]*model.User),
	}

	// Create a token manager with a fixed secret
	tokens, err := NewTokenManager(strings.Repeat("k", 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	authService := NewAuthService(fake, tokens)

	// Test invalid role
	_, err = authService.CreateUser("test@example.com", "password123", "Test User", "invalid_role")
	if err == nil {
		t.Error("Expected error for invalid role")
	}

	// Test password too short
	_, err = authService.CreateUser("test@example.com", "pass123", "Test User", "user")
	if err != ErrPasswordInvalid {
		t.Error("Expected ErrPasswordInvalid for short password")
	}

	// Test password too long
	_, err = authService.CreateUser("test@example.com", strings.Repeat("a", 73), "Test User", "user")
	if err != ErrPasswordInvalid {
		t.Error("Expected ErrPasswordInvalid for long password")
	}

	// Test valid user creation
	user, err := authService.CreateUser("Test@Example.Com", "password123", "  Test User  ", "admin")
	if err != nil {
		t.Fatal(err)
	}

	// Check that the email is lowercased and trimmed
	if user.Email != "test@example.com" {
		t.Errorf("Expected email to be lowercased and trimmed, got %s", user.Email)
	}

	// Check that the full name is trimmed
	if user.FullName != "Test User" {
		t.Errorf("Expected full name to be trimmed, got %s", user.FullName)
	}

	// Check that the role is set correctly
	if user.Role != "admin" {
		t.Errorf("Expected role to be admin, got %s", user.Role)
	}

	// Check that the password hash is not empty and not equal to the password
	if user.PasswordHash == "" {
		t.Error("Password hash should not be empty")
	}
	if user.PasswordHash == "password123" {
		t.Error("Password hash should not be equal to the password")
	}
}

func TestCreateUserEmailTaken(t *testing.T) {
	// Save original bcryptCost and restore it after the test
	originalBcryptCost := bcryptCost
	t.Cleanup(func() {
		bcryptCost = originalBcryptCost
	})
	bcryptCost = bcrypt.MinCost

	// Create a fake repository with createErr set to ErrConflict
	fake := &fakeUserRepo{
		byEmail:   make(map[string]*model.User),
		createErr: repository.ErrConflict,
	}

	// Create a token manager with a fixed secret
	tokens, err := NewTokenManager(strings.Repeat("k", 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	authService := NewAuthService(fake, tokens)

	// Try to create a user - should return ErrEmailTaken
	_, err = authService.CreateUser("test@example.com", "password123", "Test User", "user")
	if !errors.Is(err, ErrEmailTaken) {
		t.Error("Expected ErrEmailTaken when email is taken")
	}
}

func TestLoginDisabledUser(t *testing.T) {
	// Save original bcryptCost and restore it after the test
	originalBcryptCost := bcryptCost
	t.Cleanup(func() {
		bcryptCost = originalBcryptCost
	})
	bcryptCost = bcrypt.MinCost

	// Create a fake repository
	fake := &fakeUserRepo{
		byEmail: make(map[string]*model.User),
	}

	// Create a token manager with a fixed secret
	tokens, err := NewTokenManager(strings.Repeat("k", 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	authService := NewAuthService(fake, tokens)

	// Create a disabled user
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

	// Try to login with the right password - should return ErrAccountDisabled
	_, err = authService.Login("test@example.com", "password1")
	if err != ErrAccountDisabled {
		t.Error("Expected ErrAccountDisabled for disabled user")
	}

	// Try to login with the wrong password - should return ErrInvalidCredentials
	_, err = authService.Login("test@example.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Error("Expected ErrInvalidCredentials for wrong password")
	}
}
