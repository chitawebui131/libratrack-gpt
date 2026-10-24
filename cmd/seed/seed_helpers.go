package main

import (
	"crypto/rand"
	"math/big"
	"os"
	"strconv"
)

const passwordAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// randomPassword returns a random password of n letters and digits.
func randomPassword(n int) (string, error) {
	// 1. out := make([]byte, n).
	// 2. For i := range out: idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordAlphabet)))); return "", err on error; out[i] = passwordAlphabet[idx.Int64()].
	// 3. Return string(out), nil.
	out := make([]byte, n)
	for i := range out {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordAlphabet))))
		if err != nil {
			return "", err
		}
		out[i] = passwordAlphabet[idx.Int64()]
	}
	return string(out), nil
}

func envOr(name, def string) string {
	// v := os.Getenv(name); if v == "" return def; return v.
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	return v
}

func envIntOr(name string, def int) int {
	// v := os.Getenv(name); if v == "" return def. n, err := strconv.Atoi(v); if err != nil return def. Return n.
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
