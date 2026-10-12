package handler

import (
	"net/http"

	"libratrack/internal/model"
	"libratrack/internal/repository"

	"github.com/gin-gonic/gin"
)

type AuthorHandler struct {
	repo repository.AuthorRepository
}

func NewAuthorHandler(repo repository.AuthorRepository) *AuthorHandler {
	return &AuthorHandler{repo: repo}
}

func (h *AuthorHandler) Create(c *gin.Context) {
	var req model.CreateAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidation(c, err)
		return
	}
	author := model.Author{
		FullName: req.FullName,
		Bio:      req.Bio,
	}
	if err := h.repo.Create(&author); err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusCreated, author)
}

func (h *AuthorHandler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	author, err := h.repo.FindByID(id)
	if err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, author)
}

func (h *AuthorHandler) List(c *gin.Context) {
	authors, err := h.repo.List()
	if err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, authors)
}
