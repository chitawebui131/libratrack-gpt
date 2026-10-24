package service

import (
	"errors"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is a variable so that tests can lower it.
var bcryptCost = bcrypt.DefaultCost

var ErrPasswordInvalid = errors.New("password must be between 1 and 72 bytes")

// HashPassword returns a bcrypt hash of plain.
func HashPassword(plain string) (string, error) {
	// 1. If len(plain) is 0 or greater than 72, return "" and ErrPasswordInvalid.
	if len(plain) == 0 || len(plain) > 72 {
		return "", ErrPasswordInvalid
	}
	// 2. Call bcrypt.GenerateFromPassword([]byte(plain), bcryptCost).
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	// 3. Return the hash as a string and the error.
	return string(bytes), nil
}

// CheckPassword reports whether plain matches hash.
func CheckPassword(hash, plain string) bool {
	// 1. If hash is empty, return false.
	if hash == "" {
		return false
	}
	// 2. Return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil.
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}

var (
	dummyOnce sync.Once
	dummyHash []byte
)

// CheckPasswordDummy spends the same time as CheckPassword.
// It is used when the email is unknown, so response time does not reveal which emails exist.
func CheckPasswordDummy(plain string) {
	// 1. Inside dummyOnce.Do compute dummyHash with bcrypt.GenerateFromPassword([]byte("dummy-password"), bcryptCost).
	dummyOnce.Do(func() {
		var err error
		dummyHash, err = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcryptCost)
		if err != nil {
			// Handle error if generation fails
		}
	})
	// 2. Call bcrypt.CompareHashAndPassword(dummyHash, []byte(plain)) and ignore the result.
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plain))
}
