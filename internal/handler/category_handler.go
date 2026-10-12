package handler

import (
	"net/http"

	"libratrack/internal/model"
	"libratrack/internal/repository"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	repo repository.CategoryRepository
}

func NewCategoryHandler(repo repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req model.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidation(c, err)
		return
	}
	category := model.Category{Name: req.Name}
	if err := h.repo.Create(&category); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "conflict", "message": "category already exists", "details": ""}})
			return
		}
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	category, err := h.repo.FindByID(id)
	if err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.repo.List()
	if err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, categories)
}
