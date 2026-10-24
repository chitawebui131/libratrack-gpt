package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type userCreateForm struct {
	FullName string `form:"full_name" binding:"required,max=255"`
	Email    string `form:"email" binding:"required,email,max=255"`
	Password string `form:"password" binding:"required,min=8,max=72"`
	Role     string `form:"role" binding:"required,oneof=admin librarian reader"`
}

type userEditForm struct {
	FullName string `form:"full_name" binding:"required,max=255"`
	Email    string `form:"email" binding:"required,email,max=255"`
	Password string `form:"password" binding:"omitempty,min=8,max=72"`
}

// userFormPage is the data of the create/edit page. ID == 0 means create mode.
type userFormPage struct {
	ID       uint
	FullName string
	Email    string
	Role     string
	Errors   map[string]string
}

// parseUserID reads :id. It returns false for a missing, invalid or zero id.
func parseUserID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

// NewUserPage handles GET /panel/new-user.
func (h *WebPanelUsersHandler) NewUserPage(c *gin.Context) {
	h.render.Render(c, http.StatusOK, "panel_user_form", "user_new", userFormPage{Role: model.RoleReader, Errors: map[string]string{}})
}

// EditUserPage handles GET /panel/users/:id/edit.
func (h *WebPanelUsersHandler) EditUserPage(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		h.render.NotFound(c)
		return
	}
	u, err := h.users.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}
	h.render.Render(c, http.StatusOK, "panel_user_form", "user_edit", userFormPage{ID: u.ID, FullName: u.FullName, Email: u.Email, Role: u.Role, Errors: map[string]string{}})
}
