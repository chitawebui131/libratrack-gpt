package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware sets the security headers on every response.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. h := c.Writer.Header().
		h := c.Writer.Header()
		// 2. h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'").
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		// 3. h.Set("X-Content-Type-Options", "nosniff"); h.Set("X-Frame-Options", "DENY"); h.Set("Referrer-Policy", "same-origin").
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "same-origin")
		// 4. If !strings.HasPrefix(c.Request.URL.Path, "/static/"), h.Set("Cache-Control", "no-store").
		if !strings.HasPrefix(c.Request.URL.Path, "/static/") {
			h.Set("Cache-Control", "no-store")
		}
		// 5. c.Next().
		c.Next()
	}
}
