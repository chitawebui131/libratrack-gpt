package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type securityConfig struct {
	JWTSecret         string
	JWTTTL            time.Duration
	LoanPeriodDays    int
	FinePerDay        int
	FineMax           int
	MaxActiveLoans    int
	AdminEmail        string
	AdminPassword     string
	SeedDemoUsers     bool
	DemoUsersPassword string
}

// envInt reads an integer environment variable. An empty value gives def.
func envInt(name string, def int) (int, error) {
	// 1. v := os.Getenv(name). If v == "", return def, nil.
	v := os.Getenv(name)
	if v == "" {
		return def, nil
	}
	// 2. n, err := strconv.Atoi(v). If err != nil, return 0, fmt.Errorf("%s must be an integer", name).
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	// 3. Return n, nil.
	return n, nil
}
