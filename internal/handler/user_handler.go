package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
)

// UserHandler handles HTTP requests for users.
type UserHandler struct {
	repo repository.UserRepository
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// List handles GET /users (admin only).
// @Summary List users
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} []model.User
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
	// 1. page, _ := strconv.Atoi(c.Query("page")) and limit, _ := strconv.Atoi(c.Query("limit")). A value of 0 means "use the default".
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	// 2. users, err := h.repo.List(page, limit). If err != nil: respond 500 and return.
	users, err := h.repo.List(page, limit)
	if err != nil {
		respondRepoError(c, err)
		return
	}

	// 3. c.JSON(http.StatusOK, users).
	c.JSON(http.StatusOK, users)
}
