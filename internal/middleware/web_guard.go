package middleware

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// redirectToLogin sends the visitor to /login. For GET requests it adds ?next=<current URI>.
func redirectToLogin(c *gin.Context) {
	// 1. target := "/login".
	// 2. If c.Request.Method == http.MethodGet, target = "/login?next=" + url.QueryEscape(c.Request.URL.RequestURI().
	// 3. c.Redirect(http.StatusSeeOther, target); c.Abort().
	target := "/login"
	if c.Request.Method == http.MethodGet {
		target = "/login?next=" + url.QueryEscape(c.Request.URL.RequestURI())
	}
	c.Redirect(http.StatusSeeOther, target)
	c.Abort()
}

// RequireLogin redirects anonymous visitors to /login.
func RequireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If WebUser(c) is not ok: redirectToLogin(c) and return. Otherwise c.Next().
		if _, ok := WebUser(c); !ok {
			redirectToLogin(c)
			return
		}
		c.Next()
	}
}

// RequireWebRoles allows only users with one of the roles. Others get onForbidden.
func RequireWebRoles(onForbidden gin.HandlerFunc, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. u, ok := WebUser(c). If !ok: redirectToLogin(c) and return.
		// 2. Loop over roles. If u.Role equals one of them: c.Next() and return.
		// 3. onForbidden(c); c.Abort().
		u, ok := WebUser(c)
		if !ok {
			redirectToLogin(c)
			return
		}
		for _, role := range roles {
			if u.Role == role {
				c.Next()
				return
			}
		}
		onForbidden(c)
		c.Abort()
	}
}
