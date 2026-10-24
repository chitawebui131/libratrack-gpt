package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/service"
)

type registerForm struct {
	Email    string `form:"email" binding:"required,email"`
	Password string `form:"password" binding:"required,min=8,max=72"`
	FullName string `form:"full_name" binding:"required,min=2,max=255"`
}

// RegisterPage handles GET /register.
func (h *WebAuthHandler) RegisterPage(c *gin.Context) {
	// 1. If middleware.WebUser(c) is ok, redirect 303 to /catalog and return.
	if _, ok := middleware.WebUser(c); ok {
		c.Redirect(http.StatusSeeOther, "/catalog")
		return
	}
	// 2. Render page "register" with the title key "register", status 200, data authPage{Errors: map[string]string{}}.
	h.render.Render(c, http.StatusOK, "register", "register", authPage{Errors: map[string]string{}})
}

// Register handles POST /register.
func (h *WebAuthHandler) Register(c *gin.Context) {
	// 1. var f registerForm. If err := c.ShouldBind(&f); err != nil: render "register" with 422 and authPage{Email: c.PostForm("email"), FullName: c.PostForm("full_name"), Errors: fieldErrors(err)}; return.
	var f registerForm
	if err := c.ShouldBind(&f); err != nil {
		h.render.Render(c, http.StatusUnprocessableEntity, "register", "register", authPage{
			Email:    c.PostForm("email"),
			FullName: c.PostForm("full_name"),
			Errors:   fieldErrors(err),
		})
		return
	}
	// 2. _, err := h.auth.Register(model.RegisterRequest{Email: f.Email, Password: f.Password, FullName: f.FullName}).
	_, err := h.auth.Register(model.RegisterRequest{
		Email:    f.Email,
		Password: f.Password,
		FullName: f.FullName,
	})
	// 3. If errors.Is(err, service.ErrEmailTaken): render "register" with 409 and Errors{"Email": "email_taken"} (keep Email and FullName); return.
	if errors.Is(err, service.ErrEmailTaken) {
		h.render.Render(c, http.StatusConflict, "register", "register", authPage{
			Email:    f.Email,
			FullName: f.FullName,
			Errors:   map[string]string{"Email": "email_taken"},
		})
		return
	}
	// 4. If err != nil: h.render.ServerError(c); return.
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 5. resp, err := h.auth.Login(f.Email, f.Password). If err != nil: h.render.ServerError(c); return.
	resp, err := h.auth.Login(f.Email, f.Password)
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 6. middleware.SetAuthCookies(c, h.cookies, resp).
	middleware.SetAuthCookies(c, h.cookies, resp)
	// 7. RedirectWithFlash(c, "/catalog", "registered", 0).
	c.Redirect(http.StatusSeeOther, "/catalog")
}
