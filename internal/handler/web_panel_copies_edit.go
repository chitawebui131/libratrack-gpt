package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
)

type copyEditForm struct {
	InventoryNo string `form:"inventory_no" binding:"required,max=50"`
	Department  string `form:"department" binding:"max=20"`
	BookID      uint   `form:"book_id" binding:"required,gt=0"`
}

type copyDeleteForm struct {
	BookID uint `form:"book_id" binding:"required,gt=0"`
}

// UpdateCopy handles POST /panel/copies/:id.
func (h *WebPanelLoansHandler) UpdateCopy(c *gin.Context) {
	// 1. Parse :id (NotFound on error or 0). var f copyEditForm. On bind error: RedirectWithFlash(c, "/panel/books", "generic_error", 0); return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	var f copyEditForm
	if err := c.ShouldBind(&f); err != nil {
		RedirectWithFlash(c, "/panel/books", "generic_error", 0)
		return
	}
	// 2. err = h.panel.UpdateCopy(uint(id), strings.TrimSpace(f.InventoryNo), strings.TrimSpace(f.Department)).
	// 3. ErrConflict -> RedirectWithFlash(c, bookURL(f.BookID), "copy_conflict", 0); return. ErrNotFound -> NotFound; return. Other error -> ServerError; return.
	err = h.panel.UpdateCopy(uint(id), strings.TrimSpace(f.InventoryNo), strings.TrimSpace(f.Department))
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			RedirectWithFlash(c, bookURL(f.BookID), "copy_conflict", 0)
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			h.render.NotFound(c)
			return
		}
		h.render.ServerError(c)
		return
	}
	// 4. RedirectWithFlash(c, bookURL(f.BookID), "copy_updated", 0).
	RedirectWithFlash(c, bookURL(f.BookID), "copy_updated", 0)
}

// DeleteCopy handles POST /panel/copies/:id/delete.
func (h *WebPanelLoansHandler) DeleteCopy(c *gin.Context) {
	// 1. Parse :id (NotFound on error or 0). var f copyDeleteForm. On bind error: RedirectWithFlash(c, "/panel/books", "generic_error", 0); return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	var f copyDeleteForm
	if err := c.ShouldBind(&f); err != nil {
		RedirectWithFlash(c, "/panel/books", "generic_error", 0)
		return
	}
	// 2. err = h.panel.DeleteCopy(uint(id)).
	// 3. ErrInvalidReference -> RedirectWithFlash(c, bookURL(f.BookID), "copy_delete_blocked", 0); return. ErrNotFound -> NotFound; return. Other error -> ServerError; return.
	err = h.panel.DeleteCopy(uint(id))
	if err != nil {
		if errors.Is(err, repository.ErrInvalidReference) {
			RedirectWithFlash(c, bookURL(f.BookID), "copy_delete_blocked", 0)
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			h.render.NotFound(c)
			return
		}
		h.render.ServerError(c)
		return
	}
	// 4. RedirectWithFlash(c, bookURL(f.BookID), "copy_deleted", 0).
	RedirectWithFlash(c, bookURL(f.BookID), "copy_deleted", 0)
}
