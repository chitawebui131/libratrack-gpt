package handler

import (
	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
)

// Logout handles POST /logout.
func (h *WebAuthHandler) Logout(c *gin.Context) {
	// 1. refresh, err := c.Cookie(middleware.RefreshCookieName). If err == nil and refresh != "": call h.auth.Logout(refresh) and ignore its error.
	refresh, err := c.Cookie(middleware.RefreshCookieName)
	if err == nil && refresh != "" {
		h.auth.Logout(refresh)
	}
	// 2. middleware.ClearAuthCookies(c, h.cookies).
	middleware.ClearAuthCookies(c, h.cookies)
	// 3. RedirectWithFlash(c, "/login", "logged_out", 0).
	c.Redirect(303, "/login")
}
