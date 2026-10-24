package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenIssuer = "libratrack"

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time // tests may replace it
}

func NewTokenManager(secret string, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 characters")
	}
	if ttl <= 0 {
		return nil, errors.New("token ttl must be positive")
	}
	return &TokenManager{secret: []byte(secret), ttl: ttl, now: time.Now}, nil
}

func (m *TokenManager) TTL() time.Duration {
	return m.ttl
}

// Issue generates a JWT token for the given claims.
func (m *TokenManager) Issue(claims *Claims) (string, error) {
	expirationTime := m.now().Add(m.ttl)

	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	claims.Issuer = tokenIssuer

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Validate validates a JWT token string and returns the parsed claims.
func (m *TokenManager) Validate(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
