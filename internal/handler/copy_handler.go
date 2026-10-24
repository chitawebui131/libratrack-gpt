package handler

import (
	"net/http"

	"libratrack/internal/model"
	"libratrack/internal/repository"

	"github.com/gin-gonic/gin"
)

type CopyHandler struct {
	repo repository.CopyRepository
}

func NewCopyHandler(repo repository.CopyRepository) *CopyHandler {
	return &CopyHandler{repo: repo}
}

func (h *CopyHandler) Create(c *gin.Context) {
	bookID, ok := parseID(c)
	if !ok {
		return
	}
	var req model.CreateCopyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidation(c, err)
		return
	}
	copy := model.Copy{BookID: bookID, InventoryNo: req.InventoryNo, Status: req.Status}
	if err := h.repo.Create(c.Request.Context(), &copy); err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusCreated, copy)
}

func (h *CopyHandler) ListByBook(c *gin.Context) {
	bookID, ok := parseID(c)
	if !ok {
		return
	}
	copies, err := h.repo.ListByBook(c.Request.Context(), bookID)
	if err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, copies)
}
