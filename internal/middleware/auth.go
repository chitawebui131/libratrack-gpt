package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type TokenVerifier interface {
	Verify(token string) (userID uint, role string, err error)
}

// AuthMiddleware requires a valid "Authorization: Bearer <token>" header.
func AuthMiddleware(v TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2).
		parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
		// 2. If len(parts) != 2, or !strings.EqualFold(parts[0], "Bearer"), or strings.TrimSpace(parts[1]) is empty: call unauthorized(c) and return.
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			unauthorized(c)
			return
		}
		// 3. userID, role, err := v.Verify(strings.TrimSpace(parts[1])). If err != nil: call unauthorized(c) and return.
		userID, role, err := v.Verify(strings.TrimSpace(parts[1]))
		if err != nil {
			unauthorized(c)
			return
		}
		// 4. c.Set(ctxUserID, userID), c.Set(ctxRole, role), then c.Next().
		c.Set(ctxUserID, userID)
		c.Set(ctxRole, role)
		c.Next()
	}
}

// unauthorized uses one generic message for every failure, so the reason is never revealed.
// unauthorized uses one generic message for every failure, so the reason is never revealed.
func unauthorized(c *gin.Context) {

	c.Header("WWW-Authenticate", "Bearer")

	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Details string `json:"details"`
		}{
			Code:    "unauthorized",
			Message: "authentication required",
			Details: "",
		},
	})
}
