package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
	"libratrack/internal/service"
)

// CreateUser handles POST /panel/users.
func (h *WebPanelUsersHandler) CreateUser(c *gin.Context) {
	// 1. var f userCreateForm. If err := c.ShouldBind(&f); err != nil: render "panel_user_form", title "user_new", status 422 with userFormPage{FullName: f.FullName, Email: f.Email, Role: f.Role, Errors: fieldErrors(err)}; return. (Never put the password into the page data.)
	var f userCreateForm
	if err := c.ShouldBind(&f); err != nil {
		h.render.Render(c, http.StatusUnprocessableEntity, "panel_user_form", "user_new", userFormPage{FullName: f.FullName, Email: f.Email, Role: f.Role, Errors: fieldErrors(err)})
		return
	}
	// 2. _, err := h.auth.CreateUser(f.Email, f.Password, f.FullName, f.Role).
	_, err := h.auth.CreateUser(f.Email, f.Password, f.FullName, f.Role)
	// 3. errors.Is(err, service.ErrEmailTaken): render the same page with status 409 and Errors{"Email": "email_taken"}; return.
	//    errors.Is(err, service.ErrPasswordInvalid): status 422 and Errors{"Password": "password_invalid"}; return.
	//    other error: h.render.ServerError(c); return.
	if errors.Is(err, service.ErrEmailTaken) {
		h.render.Render(c, http.StatusConflict, "panel_user_form", "user_new", userFormPage{FullName: f.FullName, Email: f.Email, Role: f.Role, Errors: map[string]string{"Email": "email_taken"}})
		return
	}
	if errors.Is(err, service.ErrPasswordInvalid) {
		h.render.Render(c, http.StatusUnprocessableEntity, "panel_user_form", "user_new", userFormPage{FullName: f.FullName, Email: f.Email, Role: f.Role, Errors: map[string]string{"Password": "password_invalid"}})
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 4. RedirectWithFlash(c, "/panel/users", "user_created", 0).
	h.RedirectWithFlash(c, "/panel/users", "user_created", 0)
}

// UpdateUser handles POST /panel/users/:id. An empty password means "do not change it".
func (h *WebPanelUsersHandler) UpdateUser(c *gin.Context) {
	// 1. id, ok := parseUserID(c). If !ok: NotFound; return.
	id, ok := parseUserID(c)
	if !ok {
		h.render.NotFound(c)
		return
	}
	// 2. var f userEditForm. On bind error: render "panel_user_form", title "user_edit", status 422 with userFormPage{ID: id, FullName: f.FullName, Email: f.Email, Errors: fieldErrors(err)}; return.
	var f userEditForm
	if err := c.ShouldBind(&f); err != nil {
		h.render.Render(c, http.StatusUnprocessableEntity, "panel_user_form", "user_edit", userFormPage{ID: id, FullName: f.FullName, Email: f.Email, Errors: fieldErrors(err)})
		return
	}
	// 3. err := h.auth.UpdateProfile(id, f.Email, f.FullName). ErrEmailTaken: render the page with status 409 and Errors{"Email": "email_taken"}; return. repository.ErrNotFound: NotFound; return. Other error: ServerError; return.
	err := h.auth.UpdateProfile(id, f.Email, f.FullName)
	if errors.Is(err, service.ErrEmailTaken) {
		h.render.Render(c, http.StatusConflict, "panel_user_form", "user_edit", userFormPage{ID: id, FullName: f.FullName, Email: f.Email, Errors: map[string]string{"Email": "email_taken"}})
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 4. If f.Password != "": err = h.auth.ResetPassword(id, f.Password). service.ErrPasswordInvalid: render 422 with Errors{"Password": "password_invalid"}; return. Other error: ServerError; return. Then h.refresh.RevokeAllForUser(id) and ServerError on error.
	if f.Password != "" {
		err = h.auth.ResetPassword(id, f.Password)
		if errors.Is(err, service.ErrPasswordInvalid) {
			h.render.Render(c, http.StatusUnprocessableEntity, "panel_user_form", "user_edit", userFormPage{ID: id, FullName: f.FullName, Email: f.Email, Errors: map[string]string{"Password": "password_invalid"}})
			return
		}
		if err != nil {
			h.render.ServerError(c)
			return
		}
		err = h.refresh.RevokeAllForUser(id)
		if err != nil {
			h.render.ServerError(c)
			return
		}
	}
	// 5. RedirectWithFlash(c, "/panel/users", "user_updated", 0).
	h.RedirectWithFlash(c, "/panel/users", "user_updated", 0)
}
