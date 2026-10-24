package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/service"
)

// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Login request"
// @Success 200 {object} model.AuthResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	// 1. var req model.LoginRequest. If c.ShouldBindJSON(&req) fails: respond 422 and return.
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidation(c, err)
		return
	}
	// 2. resp, err := h.svc.Login(req.Email, req.Password).
	resp, err := h.svc.Login(req.Email, req.Password)
	// 3. If errors.Is(err, service.ErrInvalidCredentials): respond 401 with code "unauthorized" and message "invalid email or password", and return.
	if errors.Is(err, service.ErrInvalidCredentials) {
		respondError(c, http.StatusUnauthorized, "unauthorized", "invalid email or password", "")
		return
	}
	// 4. If err != nil: respond 500 and return.
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal_error", err.Error(), "")
		return
	}
	// 5. c.JSON(http.StatusOK, resp).
	c.JSON(http.StatusOK, resp)
}

// @Summary Get current user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} model.User

// @Router /me [get]
// @Security BearerAuth
func (h *AuthHandler) Me(c *gin.Context) {
	// 1. userID, _, ok := middleware.CurrentUser(c). If !ok: respond 401 and return.
	userID, _, ok := middleware.CurrentUser(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "unauthorized", "")
		return
	}
	// 2. user, err := h.svc.GetByID(userID). If err != nil: respond 401 and return.
	user, err := h.svc.GetByID(userID)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "unauthorized", "unauthorized", "")
		return
	}
	// 3. c.JSON(http.StatusOK, user).
	c.JSON(http.StatusOK, user)
}
