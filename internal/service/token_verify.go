package service

import (
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v5"

	"libratrack/internal/model"
)

// Verify checks the token and returns the user id and role.
// Every failure wraps ErrInvalidToken. The token text is never put into an error.
func (m *TokenManager) Verify(token string) (uint, string, error) {
	// 1. claims := &Claims{}
	claims := &Claims{}
	// 2. _, err := jwt.ParseWithClaims(token, claims,
	//        func(t *jwt.Token) (interface{}, error) { return m.secret, nil },
	//        jwt.WithValidMethods([]string{"HS256"}),
	//        jwt.WithIssuer(tokenIssuer),
	//        jwt.WithExpirationRequired(),
	//        jwt.WithTimeFunc(m.now))
	_, err := jwt.ParseWithClaims(token, claims,
		func(t *jwt.Token) (interface{}, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now))
	// 3. If err != nil, return 0, "", fmt.Errorf("%w: %v", ErrInvalidToken, err).
	if err != nil {
		return 0, "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	// 4. id, err := strconv.ParseUint(claims.Subject, 10, 64). If err != nil or id == 0, return 0, "", ErrInvalidToken.
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || id == 0 {
		return 0, "", ErrInvalidToken
	}
	// 5. If !model.IsValidRole(claims.Role), return 0, "", ErrInvalidToken.
	if !model.IsValidRole(claims.Role) {
		return 0, "", ErrInvalidToken
	}
	// 6. Return uint(id), claims.Role, nil.
	return uint(id), claims.Role, nil
}
