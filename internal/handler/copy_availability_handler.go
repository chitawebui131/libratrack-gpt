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

func (h *CopyHandler) Availability(c *gin.Context) {
	bookID, ok := parseID(c)
	if !ok {
		return
	}
	result, err := h.repo.Availability(bookID)
	if err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
