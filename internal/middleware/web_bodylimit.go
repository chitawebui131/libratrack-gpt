package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimitMiddleware limits the size of every request body to max bytes.
func BodyLimitMiddleware(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max).
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		// 2. c.Next().
		c.Next()
	}
}
