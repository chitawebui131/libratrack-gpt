package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		log.Printf("%s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered: %v", r)
				c.AbortWithStatusJSON(500, gin.H{
					"error": gin.H{
						"code":    "internal_error",
						"message": "internal server error",
						"details": "",
					},
				})
			}
		}()
		c.Next()
	}
}
