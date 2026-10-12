package handler

import (
	"net/http"

	"libratrack/internal/repository"

	"github.com/gin-gonic/gin"
)

type CopyAvailabilityHandler struct {
	repo repository.CopyRepository
}

func NewCopyAvailabilityHandler(repo repository.CopyRepository) *CopyAvailabilityHandler {
	return &CopyAvailabilityHandler{repo: repo}
}

func (h *CopyAvailabilityHandler) Availability(c *gin.Context) {
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
