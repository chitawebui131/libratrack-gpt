package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"libratrack/internal/model"
)

func (h *BookHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req model.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidation(c, err)
		return
	}
	book := model.Book{
		ID:            id,
		Title:         req.Title,
		ISBN:          req.ISBN,
		CategoryID:    req.CategoryID,
		Publisher:     req.Publisher,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	}
	if err := h.repo.Update(&book, req.AuthorIDs); err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, book)
}

func (h *BookHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.repo.Delete(id); err != nil {
		respondRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
