package service

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func init() {
	bcryptCost = bcrypt.MinCost
}

func TestHashPasswordDiffersFromPlain(t *testing.T) {
	// Hash "secret123". The hash must be non-empty and different from "secret123".
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash is empty")
	}

	if hash == "secret123" {
		t.Error("Hash equals plain text")
	}
}

func TestCheckPasswordRightAndWrong(t *testing.T) {
	// Hash "secret123". CheckPassword with "secret123" is true. With "wrong-pass" it is false.
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !CheckPassword(hash, "secret123") {
		t.Error("CheckPassword should return true for correct password")
	}

	if CheckPassword(hash, "wrong-pass") {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestCheckPasswordEmptyHash(t *testing.T) {
	// CheckPassword("", "secret123") is false.
	if CheckPassword("", "secret123") {
		t.Error("CheckPassword with empty hash should return false")
	}
}

func TestHashPasswordInvalid(t *testing.T) {
	// HashPassword("") and HashPassword(strings.Repeat("a", 73)) must return an error that satisfies errors.Is(err, ErrPasswordInvalid).
	_, err := HashPassword("")
	if err == nil {
		t.Error("HashPassword with empty string should return an error")
	}
	if !errors.Is(err, ErrPasswordInvalid) {
		t.Errorf("Expected ErrPasswordInvalid, got %v", err)
	}

	_, err = HashPassword(strings.Repeat("a", 73))
	if err == nil {
		t.Error("HashPassword with 73 characters should return an error")
	}
	if !errors.Is(err, ErrPasswordInvalid) {
		t.Errorf("Expected ErrPasswordInvalid, got %v", err)
	}
}
