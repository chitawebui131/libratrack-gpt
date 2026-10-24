package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
	_ "time/tzdata"

	"libratrack/internal/middleware"
)

type webConfig struct {
	Location *time.Location
	Cookies  middleware.CookieConfig
}

func loadWebConfig() (webConfig, error) {
	// 1. secure := false. If v := os.Getenv("COOKIE_SECURE"); v != "": b, err := strconv.ParseBool(v); on error return webConfig{}, fmt.Errorf("COOKIE_SECURE must be true or false"); secure = b.
	secure := false
	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return webConfig{}, fmt.Errorf("COOKIE_SECURE must be true or false")
		}
		secure = b
	}

	// 2. tz := os.Getenv("APP_TIMEZONE"); if tz == "" { tz = "UTC" }. loc, err := time.LoadLocation(tz). On error return webConfig{}, fmt.Errorf("invalid APP_TIMEZONE: %v", err).
	tz := os.Getenv("APP_TIMEZONE")
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return webConfig{}, fmt.Errorf("invalid APP_TIMEZONE: %v", err)
	}

	// 3. ttlMinutes, err := envInt("JWT_TTL_MINUTES", 60) and refreshHours, err := envInt("REFRESH_TTL_HOURS", 168). On error return webConfig{}, err.
	ttlMinutes, err := envInt("JWT_TTL_MINUTES", 60)
	if err != nil {
		return webConfig{}, err
	}
	refreshHours, err := envInt("REFRESH_TTL_HOURS", 168)
	if err != nil {
		return webConfig{}, err
	}

	// 4. Return webConfig{Location: loc, Cookies: middleware.CookieConfig{Secure: secure, AccessMaxAge: ttlMinutes * 60, RefreshMaxAge: refreshHours * 3600}}, nil.
	return webConfig{
		Location: loc,
		Cookies:  middleware.CookieConfig{Secure: secure, AccessMaxAge: ttlMinutes * 60, RefreshMaxAge: refreshHours * 3600},
	}, nil
}
