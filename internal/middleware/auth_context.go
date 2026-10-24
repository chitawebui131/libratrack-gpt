package middleware

import "github.com/gin-gonic/gin"

const (
	ctxUserID = "auth_user_id"
	ctxRole   = "auth_role"
)

// CurrentUser returns the user id and role stored by AuthMiddleware.
func CurrentUser(c *gin.Context) (userID uint, role string, ok bool) {
	// 1. idValue, exists := c.Get(ctxUserID). If !exists, return 0, "", false.
	idValue, exists := c.Get(ctxUserID)
	if !exists {
		return 0, "", false
	}
	// 2. roleValue, exists := c.Get(ctxRole). If !exists, return 0, "", false.
	roleValue, exists := c.Get(ctxRole)
	if !exists {
		return 0, "", false
	}
	// 3. Convert idValue to uint and roleValue to string with type assertions. If a conversion fails, return 0, "", false.
	id, ok := idValue.(uint)
	if !ok {
		return 0, "", false
	}
	role, ok = roleValue.(string)
	if !ok {
		return 0, "", false
	}
	// 4. Return the id, the role and true.
	return id, role, true
}

// abortError stops the request with the standard error envelope.
func abortError(c *gin.Context, status int, code, message string) {
	// Call c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message, "details": ""}}).
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message, "details": ""}})
}
