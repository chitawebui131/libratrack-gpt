package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/service"
)

// Login handles POST /login.
func (h *WebAuthHandler) Login(c *gin.Context) {
	// 1. next := safeNext(c.PostForm("next")).
	next := safeNext(c.PostForm("next"))
	// 2. var f loginForm. If err := c.ShouldBind(&f); err != nil: render page "login" with status 422 and authPage{Email: c.PostForm("email"), Next: next, Errors: fieldErrors(err)}; return.
	var f loginForm
	if err := c.ShouldBind(&f); err != nil {
		h.render.Render(c, http.StatusUnprocessableEntity, "login", "login", authPage{Email: c.PostForm("email"), Next: next, Errors: fieldErrors(err)})
		return
	}
	// 3. resp, err := h.auth.Login(f.Email, f.Password).
	resp, err := h.auth.Login(f.Email, f.Password)
	// 4. If errors.Is(err, service.ErrInvalidCredentials): render "login" with status 401 and authPage{Email: f.Email, Next: next, Errors: map[string]string{"Password": "invalid_credentials"}}; return.
	if errors.Is(err, service.ErrInvalidCredentials) {
		h.render.Render(c, http.StatusUnauthorized, "login", "login", authPage{Email: f.Email, Next: next, Errors: map[string]string{"Password": "invalid_credentials"}})
		return
	}
	// 5. If errors.Is(err, service.ErrAccountDisabled): render "login" with status 403 and authPage{Email: f.Email, Next: next, Errors: map[string]string{"Password": "account_disabled"}}; return.
	if errors.Is(err, service.ErrAccountDisabled) {
		h.render.Render(c, http.StatusForbidden, "login", "login", authPage{Email: f.Email, Next: next, Errors: map[string]string{"Password": "account_disabled"}})
		return
	}
	// 6. If err != nil: h.render.ServerError(c); return.
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 7. middleware.SetAuthCookies(c, h.cookies, resp).
	middleware.SetAuthCookies(c, h.cookies, resp)
	// 8. Staff land on the panel unless a specific page was requested.
	if next == "/catalog" && (resp.User.Role == model.RoleLibrarian || resp.User.Role == model.RoleAdmin) {
		next = "/panel"
	}
	// 9. c.Redirect(http.StatusSeeOther, next).
	c.Redirect(http.StatusSeeOther, next)
}
