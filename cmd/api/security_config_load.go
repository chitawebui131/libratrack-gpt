package main

import (
	"errors"
	"os"
	"time"
)

func loadSecurityConfig() (securityConfig, error) {
	var cfg securityConfig
	// 1. cfg.JWTSecret = os.Getenv("JWT_SECRET"). If len(cfg.JWTSecret) < 32, return cfg and errors.New("JWT_SECRET must be set and at least 32 characters").
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < 32 {
		return cfg, errors.New("JWT_SECRET must be set and at least 32 characters")
	}
	// 2. ttlMinutes, err := envInt("JWT_TTL_MINUTES", 60). On error return cfg, err. Set cfg.JWTTTL = time.Duration(ttlMinutes) * time.Minute.
	ttlMinutes, err := envInt("JWT_TTL_MINUTES", 60)
	if err != nil {
		return cfg, err
	}
	cfg.JWTTTL = time.Duration(ttlMinutes) * time.Minute
	// 3. In the same way read with envInt: LOAN_PERIOD_DAYS (default 14) -> cfg.LoanPeriodDays, FINE_PER_DAY (default 2) -> cfg.FinePerDay,
	//    FINE_MAX (default 100) -> cfg.FineMax, MAX_ACTIVE_LOANS (default 5) -> cfg.MaxActiveLoans.
	cfg.LoanPeriodDays, err = envInt("LOAN_PERIOD_DAYS", 14)
	if err != nil {
		return cfg, err
	}
	cfg.FinePerDay, err = envInt("FINE_PER_DAY", 2)
	if err != nil {
		return cfg, err
	}
	cfg.FineMax, err = envInt("FINE_MAX", 100)
	if err != nil {
		return cfg, err
	}
	cfg.MaxActiveLoans, err = envInt("MAX_ACTIVE_LOANS", 5)
	if err != nil {
		return cfg, err
	}
	// 4. cfg.AdminEmail = os.Getenv("ADMIN_EMAIL"), cfg.AdminPassword = os.Getenv("ADMIN_PASSWORD").
	cfg.AdminEmail = os.Getenv("ADMIN_EMAIL")
	cfg.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	// 5. cfg.SeedDemoUsers = os.Getenv("SEED_DEMO_USERS") == "true", cfg.DemoUsersPassword = os.Getenv("DEMO_USERS_PASSWORD").
	cfg.SeedDemoUsers = os.Getenv("SEED_DEMO_USERS") == "true"
	cfg.DemoUsersPassword = os.Getenv("DEMO_USERS_PASSWORD")
	// 6. Return cfg, nil.
	return cfg, nil
}
