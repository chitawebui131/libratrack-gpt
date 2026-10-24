package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRoles allows only users whose role is in the list.
// It must run after AuthMiddleware.
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. _, role, ok := CurrentUser(c). If !ok: call unauthorized(c) and return.
		_, role, ok := CurrentUser(c)
		if !ok {
			unauthorized(c)
			return
		}
		// 2. Loop over roles. If role equals one of them: call c.Next() and return.
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}
		// 3. abortError(c, http.StatusForbidden, "forbidden", "insufficient permissions").
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": struct {
				Code    string `json:"code"`
				Message string `json:"message"`
				Details string `json:"details"`
			}{
				Code:    "forbidden",
				Message: "insufficient permissions",
				Details: "",
			},
		})

	}
}
