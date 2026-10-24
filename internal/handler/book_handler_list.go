package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func queryInt(c *gin.Context, name string, def int) (int, bool) {
	s := c.Query(name)
	if s == "" {
		return def, true
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		respondError(c, http.StatusUnprocessableEntity, "validation_error", "invalid query parameter", name)
		return 0, false
	}
	return v, true
}

func (h *BookHandler) List(c *gin.Context) {
	categoryID, ok := queryInt(c, "category_id", 0)
	if !ok {
		return
	}
	page, ok := queryInt(c, "page", 1)
	if !ok {
		return
	}
	limit, ok := queryInt(c, "limit", 20)
	if !ok {
		return
	}
	books, err := h.repo.List(uint(categoryID), page, limit)
	if err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, books)
}
