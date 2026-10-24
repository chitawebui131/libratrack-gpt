package service

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Issue creates a signed HS256 token and returns it together with its expiry time.
func (m *TokenManager) IssueWithNewName(userID uint, role string) (string, time.Time, error) {
	now := m.now().UTC()
	expires := now.Add(m.ttl)
	claims := Claims{Role: role, RegisteredClaims: jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		Issuer:    tokenIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expires),
	}}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return signed, expires, err
}
