package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
)

// DeleteAuthor handles POST /panel/authors/:id/delete.
func (h *WebPanelDictionariesHandler) DeleteAuthor(c *gin.Context) {
	// 1. id, err := strconv.ParseUint(c.Param("id"), 10, 32). If err != nil or id == 0: h.render.NotFound(c); return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	// 2. err = h.dict.DeleteAuthor(uint(id)).
	err = h.dict.DeleteAuthor(uint(id))
	// 3. errors.Is(err, repository.ErrNotFound): h.render.NotFound(c); return.
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	// 4. errors.Is(err, repository.ErrInvalidReference): RedirectWithFlash(c, "/panel/authors", "author_delete_blocked", 0); return.
	if errors.Is(err, repository.ErrInvalidReference) {
		RedirectWithFlash(c, "/panel/authors", "author_delete_blocked", 0)
		return
	}
	// 5. Other error: h.render.ServerError(c); return.
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 6. RedirectWithFlash(c, "/panel/authors", "author_deleted", 0).
	RedirectWithFlash(c, "/panel/authors", "author_deleted", 0)
}

// DeleteCategory handles POST /panel/categories/:id/delete. Same logic with h.dict.DeleteCategory, redirects to /panel/categories and the flash codes category_delete_blocked and category_deleted.
func (h *WebPanelDictionariesHandler) DeleteCategory(c *gin.Context) {
	// Same six steps as DeleteAuthor.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	err = h.dict.DeleteCategory(uint(id))
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	if errors.Is(err, repository.ErrInvalidReference) {
		RedirectWithFlash(c, "/panel/categories", "category_delete_blocked", 0)
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}
	RedirectWithFlash(c, "/panel/categories", "category_deleted", 0)
}
