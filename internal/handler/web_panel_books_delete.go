package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
)

// DeleteBook handles POST /panel/books/:id/delete.
func (h *WebPanelBooksHandler) DeleteBook(c *gin.Context) {
	// 1. id, err := strconv.ParseUint(c.Param("id"), 10, 32). If err != nil or id == 0: h.render.NotFound(c); return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	// 2. err = h.panel.DeleteBook(uint(id)).
	err = h.panel.DeleteBook(uint(id))
	// 3. If errors.Is(err, repository.ErrNotFound): h.render.NotFound(c); return.
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	// 4. If errors.Is(err, repository.ErrInvalidReference): RedirectWithFlash(c, "/books/<id>", "book_delete_blocked", 0); return.
	if errors.Is(err, repository.ErrInvalidReference) {
		RedirectWithFlash(c, "/books/"+strconv.FormatUint(id, 10), "book_delete_blocked", 0)
		return
	}
	// 5. If err != nil: h.render.ServerError(c); return.
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 6. RedirectWithFlash(c, "/panel/books", "book_deleted", 0).
	RedirectWithFlash(c, "/panel/books", "book_deleted", 0)
}
