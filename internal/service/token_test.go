package service

import (
	"errors"
	"testing"
	"time"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func newTestManager(t *testing.T) *TokenManager {
	// Create a manager with testSecret and a ttl of one hour using NewTokenManager. Fail the test on error. Return it.
	m, err := NewTokenManager(testSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}
	return m
}

func TestNewTokenManagerRejectsBadInput(t *testing.T) {
	// NewTokenManager("short", time.Hour) must return an error. NewTokenManager(testSecret, 0) must return an error.
	_, err := NewTokenManager("short", time.Hour)
	if err == nil {
		t.Error("NewTokenManager with short secret should return an error")
	}
	_, err = NewTokenManager(testSecret, 0)
	if err == nil {
		t.Error("NewTokenManager with zero ttl should return an error")
	}
}

func TestIssueThenVerify(t *testing.T) {
	// Issue a token for user id 7 and role "librarian". Verify it. The id must be 7 and the role "librarian".
	m := newTestManager(t)
	token, _, err := m.IssueWithNewName(7, "librarian")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}
	id, role, err := m.Verify(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}
	if id != 7 {
		t.Errorf("Expected id 7, got %d", id)
	}
	if role != "librarian" {
		t.Errorf("Expected role librarian, got %s", role)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	// Issue a token. Then set m.now to a function returning time.Now().Add(2 * time.Hour). Verify must return an error that satisfies errors.Is(err, ErrInvalidToken).
	m := newTestManager(t)
	token, _, err := m.IssueWithNewName(1, "patron")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}
	// Override now function to return a time in the past
	m.now = func() time.Time {
		return time.Now().Add(2 * time.Hour)
	}
	_, _, err = m.Verify(token)
	if err == nil {
		t.Error("Expected error when verifying expired token")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}
