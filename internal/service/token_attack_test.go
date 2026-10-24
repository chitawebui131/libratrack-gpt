package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyWrongSecret(t *testing.T) {
	// Issue with newTestManager. Create a second manager with another 32-character secret. Its Verify must fail with ErrInvalidToken.
	m1 := newTestManager(t)
	token, _, err := m1.IssueWithNewName(1, "reader")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	m2, err := NewTokenManager("abcdefghijklmnopqrstuvwx1234567890", time.Hour)
	if err != nil {
		t.Fatalf("Failed to create second token manager: %v", err)
	}

	_, _, err = m2.Verify(token)
	if err == nil {
		t.Error("Expected error when verifying with wrong secret")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyTamperedToken(t *testing.T) {
	// Issue a token. Change its last character. Verify must fail with ErrInvalidToken.
	m := newTestManager(t)
	token, _, err := m.IssueWithNewName(1, "reader")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// Tamper with the token by changing its last character
	tamperedToken := token[:len(token)-1] + "X"

	_, _, err = m.Verify(tamperedToken)
	if err == nil {
		t.Error("Expected error when verifying tampered token")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyNoneAlgorithm(t *testing.T) {
	// Build a token with jwt.NewWithClaims(jwt.SigningMethodNone, Claims{...valid subject "1", role "reader", issuer tokenIssuer, expiry in the future...}).
	// Sign it with jwt.UnsafeAllowNoneSignatureType (test only). Verify must fail with ErrInvalidToken.
	expirationTime := time.Now().Add(time.Hour)

	claims := &Claims{
		Role: "reader",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    tokenIssuer,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	// Sign with UnsafeAllowNoneSignatureType for testing purposes only
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	m := newTestManager(t)
	_, _, err = m.Verify(tokenString)
	if err == nil {
		t.Error("Expected error when verifying token with none algorithm")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyWrongIssuer(t *testing.T) {
	// Build and sign (HS256, testSecret) a token with issuer "someone-else". Verify must fail with ErrInvalidToken.
	expirationTime := time.Now().Add(time.Hour)

	claims := &Claims{
		Role: "reader",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "someone-else",
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	m := newTestManager(t)
	_, _, err = m.Verify(tokenString)
	if err == nil {
		t.Error("Expected error when verifying token with wrong issuer")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyUnknownRole(t *testing.T) {
	// Build and sign (HS256, testSecret) a token with role "superuser" and the correct issuer. Verify must fail with ErrInvalidToken.
	expirationTime := time.Now().Add(time.Hour)

	claims := &Claims{
		Role: "superuser",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    tokenIssuer,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	m := newTestManager(t)
	_, _, err = m.Verify(tokenString)
	if err == nil {
		t.Error("Expected error when verifying token with unknown role")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyErrorDoesNotContainToken(t *testing.T) {
	// Issue a token and change its last character. The text of the Verify error must not contain the token string (use strings.Contains).
	m := newTestManager(t)
	token, _, err := m.IssueWithNewName(1, "reader")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// Tamper with the token by changing its last character
	tamperedToken := token[:len(token)-1] + "X"

	_, _, err = m.Verify(tamperedToken)
	if err == nil {
		t.Error("Expected error when verifying tampered token")
	}
	if strings.Contains(err.Error(), token) {
		t.Errorf("Error message should not contain the token string")
	}
}
