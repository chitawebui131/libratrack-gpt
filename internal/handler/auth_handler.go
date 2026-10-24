package handler

import (
	"errors"
	"libratrack/internal/model"
	"libratrack/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Register(req model.RegisterRequest) (*model.User, error)
	Login(email, password string) (*model.AuthResponse, error)
	GetByID(id uint) (*model.User, error)
}

type AuthHandler struct {
	svc AuthService
}

func NewAuthHandler(svc AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body model.RegisterRequest true "Register request"
// @Success 201 {object} model.User
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidation(c, err)
		return
	}

	user, err := h.svc.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			respondError(c, http.StatusConflict, "email_taken", err.Error(), "")
			return
		}
		respondError(c, http.StatusInternalServerError, "internal_error", err.Error(), "")
		return
	}

	c.JSON(http.StatusCreated, user)
}
