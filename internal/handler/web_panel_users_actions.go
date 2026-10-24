package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/repository"
)

type userDisableForm struct {
	Disabled string `form:"disabled" binding:"required,oneof=0 1"`
}

// SetDisabled handles POST /panel/users/:id/disable (form field disabled = "1" or "0").
func (h *WebPanelUsersHandler) SetDisabled(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var f userDisableForm
	if err := c.ShouldBind(&f); err != nil {
		RedirectWithFlash(c, "/panel/users", "generic_error", 0)
		return
	}

	current, _ := middleware.WebUser(c)
	if current != nil && current.ID == id {
		RedirectWithFlash(c, "/panel/users", "user_self_action", 0)
		return
	}

	disabled := f.Disabled == "1"
	err := h.users.SetDisabled(id, disabled)
	if errors.Is(err, repository.ErrNotFound) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if disabled {
		err = h.refresh.RevokeAllForUser(id)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	code := "user_enabled"
	if disabled {
		code = "user_disabled"
	}
	RedirectWithFlash(c, "/panel/users", code, 0)
}

// DeleteUser handles POST /panel/users/:id/delete.
func (h *WebPanelUsersHandler) DeleteUser(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	current, _ := middleware.WebUser(c)
	if current != nil && current.ID == id {
		RedirectWithFlash(c, "/panel/users", "user_self_action", 0)
		return
	}

	err := h.users.Delete(id)
	if errors.Is(err, repository.ErrNotFound) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if errors.Is(err, repository.ErrInvalidReference) {
		RedirectWithFlash(c, "/panel/users", "user_delete_blocked", 0)
		return
	}
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	RedirectWithFlash(c, "/panel/users", "user_deleted", 0)
}
