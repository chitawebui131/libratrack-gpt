package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
)

// WebAuthService is implemented by *service.AuthService.
type WebAuthService interface {
	Register(req model.RegisterRequest) (*model.User, error)
	Login(email, password string) (*model.AuthResponse, error)
	Logout(refreshToken string) error
}

type WebAuthHandler struct {
	auth    WebAuthService
	render  *Renderer
	cookies middleware.CookieConfig
}

func NewWebAuthHandler(auth WebAuthService, render *Renderer, cookies middleware.CookieConfig) *WebAuthHandler {
	return &WebAuthHandler{auth: auth, render: render, cookies: cookies}
}

type loginForm struct {
	Email    string `form:"email" binding:"required,email"`
	Password string `form:"password" binding:"required"`
}

// authPage is the data of the login and register pages.
type authPage struct {
	Email    string
	FullName string
	Next     string
	Errors   map[string]string
}

// Index redirects / to /catalog.
func (h *WebAuthHandler) Index(c *gin.Context) {
	if u, ok := middleware.WebUser(c); ok && (u.Role == model.RoleLibrarian || u.Role == model.RoleAdmin) {
		c.Redirect(http.StatusSeeOther, "/panel")
		return
	}
	c.Redirect(http.StatusSeeOther, "/catalog")
}

// LoginPage handles GET /login.
func (h *WebAuthHandler) LoginPage(c *gin.Context) {
	// 1. If middleware.WebUser(c) is ok, redirect 303 to /catalog and return.
	if _, ok := middleware.WebUser(c); ok {
		u, _ := middleware.WebUser(c)
		if u.Role == model.RoleLibrarian || u.Role == model.RoleAdmin {
			c.Redirect(http.StatusSeeOther, "/panel")
		} else {
			c.Redirect(http.StatusSeeOther, "/catalog")
		}
		return
	}
	// 2. h.render.Render(c, http.StatusOK, "login", "login", authPage{Next: safeNext(c.Query("next")), Errors: map[string]string{}}).
	h.render.Render(c, http.StatusOK, "login", "login", authPage{Next: safeNext(c.Query("next")), Errors: map[string]string{}})
}
