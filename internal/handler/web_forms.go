package handler

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// fieldErrors maps the failed fields of a validator error to the failed validation tag.
func fieldErrors(err error) map[string]string {
	out := map[string]string{}
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range errs {
			out[fe.StructField()] = fe.Tag()
		}
	}
	return out
}

// safeNext returns next only when it is a local path. Otherwise "/catalog" (prevents open redirects).
func safeNext(next string) string {
	if next == "" {
		return "/catalog"
	}
	if next[0] != '/' {
		return "/catalog"
	}
	if len(next) > 1 && (next[1] == '/' || next[1] == '\\') {
		return "/catalog"
	}
	if strings.ContainsAny(next, "\r\n") {
		return "/catalog"
	}
	u, err := url.Parse(next)
	if err != nil || u.Scheme != "" || u.Host != "" {
		return "/catalog"
	}
	return next
}

// parseUintParam returns 0 for an empty or invalid value.
func parseUintParam(s string) uint {
	n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0
	}
	return uint(n)
}

// parsePage returns the page number, at least 1.
func parsePage(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 {
		return 1
	}
	return n
}
