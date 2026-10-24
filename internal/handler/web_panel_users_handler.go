package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/repository"
)

// UserAdmin is implemented by *service.AuthService.
type UserAdmin interface {
	CreateUser(email, password, fullName, role string) (*model.User, error)
	UpdateProfile(id uint, email, fullName string) error
	ResetPassword(id uint, password string) error
}

type WebPanelUsersHandler struct {
	users   repository.UserRepository
	refresh repository.RefreshTokenRepository
	auth    UserAdmin
	render  *Renderer
}

func NewWebPanelUsersHandler(users repository.UserRepository, refresh repository.RefreshTokenRepository, auth UserAdmin, render *Renderer) *WebPanelUsersHandler {
	return &WebPanelUsersHandler{users: users, refresh: refresh, auth: auth, render: render}
}

type roleForm struct {
	Role string `form:"role" binding:"required,oneof=admin librarian reader"`
}

type usersPage struct {
	Users []model.User
	Pager Pager
}

// UsersPage handles GET /panel/users (20 per page).
func (h *WebPanelUsersHandler) UsersPage(c *gin.Context) {
	// 1. page := parsePage(c.Query("page")); const limit = 20.
	page := parsePage(c.Query("page"))
	const limit = 20
	// 2. users, err := h.users.List(page, limit). On error: h.render.ServerError(c); return.
	users, err := h.users.List(page, limit)
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 3. total := int64((page-1)*limit + len(users)). If len(users) == limit, total++ (so that a "next" link appears when the page is full).
	total := int64((page-1)*limit + len(users))
	if len(users) == limit {
		total++
	}
	// 4. h.render.Render(c, http.StatusOK, "panel_users", "users", usersPage{Users: users, Pager: BuildPager(c.Request, page, limit, total)}).
	h.render.Render(c, http.StatusOK, "panel_users", "users", usersPage{Users: users, Pager: BuildPager(c.Request, page, limit, total)})
}

// ChangeRole handles POST /panel/users/:id/role.
func (h *WebPanelUsersHandler) ChangeRole(c *gin.Context) {
	// 1. id, err := strconv.ParseUint(c.Param("id"), 10, 32). If err != nil or id == 0: h.render.NotFound(c); return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	// 2. var f roleForm. If err := c.ShouldBind(&f); err != nil: RedirectWithFlash(c, "/panel/users", "generic_error", 0); return.
	var f roleForm
	if err := c.ShouldBind(&f); err != nil {
		RedirectWithFlash(c, "/panel/users", "generic_error", 0)
		return
	}
	// 3. current, _ := middleware.WebUser(c). If current != nil and current.ID == uint(id): RedirectWithFlash(c, "/panel/users", "role_self_change", 0); return.
	current, _ := middleware.WebUser(c)
	if current != nil && current.ID == uint(id) {
		RedirectWithFlash(c, "/panel/users", "role_self_change", 0)
		return
	}
	// 4. _, err = h.users.UpdateRole(uint(id), f.Role). If errors.Is(err, repository.ErrNotFound): h.render.NotFound(c); return. Other error: ServerError; return.
	_, err = h.users.UpdateRole(uint(id), f.Role)
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 5. If err := h.refresh.RevokeAllForUser(uint(id)); err != nil: h.render.ServerError(c); return.
	if err := h.refresh.RevokeAllForUser(uint(id)); err != nil {
		h.render.ServerError(c)
		return
	}
	// 6. RedirectWithFlash(c, "/panel/users", "role_updated", 0).
	RedirectWithFlash(c, "/panel/users", "role_updated", 0)
}
