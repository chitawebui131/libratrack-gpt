package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/repository"
)

// UpdateRole handles PUT /users/:id/role (admin only).
// @Summary Update user role
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body model.UpdateUserRoleRequest true "Role update request"
// @Success 200 {object} model.User
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{id}/role [put]
func (h *UserHandler) UpdateRole(c *gin.Context) {
	// 1. Parse the id from c.Param("id"). If it is invalid: respond 422 and return.
	id, ok := parseID(c)
	if !ok {
		return
	}

	// 2. var req model.UpdateUserRoleRequest. If c.ShouldBindJSON(&req) fails: respond 422 and return.
	var req model.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidation(c, err)
		return
	}

	// 3. currentID, _, _ := middleware.CurrentUser(c). If currentID == id: respond 403, code "forbidden", message "you cannot change your own role", and return.
	currentID, _, ok := middleware.CurrentUser(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "unauthorized", "")
		return
	}
	if currentID == id {
		respondError(c, http.StatusForbidden, "forbidden", "you cannot change your own role", "")
		return
	}

	// 4. user, err := h.repo.UpdateRole(id, req.Role).
	user, err := h.repo.UpdateRole(id, req.Role)

	// 5. If errors.Is(err, repository.ErrNotFound): respond 404 and return. Any other error: respond 500 and return.
	if errors.Is(err, repository.ErrNotFound) {
		respondRepoError(c, err)
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal_error", "internal error", err.Error())
		return
	}

	// 6. c.JSON(http.StatusOK, user).
	c.JSON(http.StatusOK, user)
}
