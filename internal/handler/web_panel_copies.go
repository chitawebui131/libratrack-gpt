package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
)

type copyForm struct {
	InventoryNo string `form:"inventory_no" binding:"required,max=50"`
	Department  string `form:"department" binding:"max=20"`
}

type copyStatusForm struct {
	Status string `form:"status" binding:"required,oneof=available lost under_repair"`
	BookID uint   `form:"book_id" binding:"required,gt=0"`
}

func bookURL(id uint) string {
	return "/books/" + strconv.FormatUint(uint64(id), 10)
}

// AddCopy handles POST /panel/books/:id/copies.
func (h *WebPanelLoansHandler) AddCopy(c *gin.Context) {
	// 1. id, err := strconv.ParseUint(c.Param("id"), 10, 32). If err != nil or id == 0: h.render.NotFound(c); return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	// 2. var f copyForm. If err := c.ShouldBind(&f); err != nil: RedirectWithFlash(c, bookURL(uint(id)), "generic_error", 0); return.
	var f copyForm
	if err := c.ShouldBind(&f); err != nil {
		RedirectWithFlash(c, bookURL(uint(id)), "generic_error", 0)
		return
	}
	// 3. err = h.panel.AddCopy(uint(id), strings.TrimSpace(f.InventoryNo), strings.TrimSpace(f.Department)).
	// 4. ErrNotFound -> h.render.NotFound(c); return. ErrConflict -> RedirectWithFlash(c, bookURL(uint(id)), "generic_error", 0); return. Other error -> h.render.ServerError(c); return.
	err = h.panel.AddCopy(uint(id), strings.TrimSpace(f.InventoryNo), strings.TrimSpace(f.Department))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.render.NotFound(c)
			return
		}
		if errors.Is(err, repository.ErrConflict) {
			RedirectWithFlash(c, bookURL(uint(id)), "generic_error", 0)
			return
		}
		h.render.ServerError(c)
		return
	}
	// 5. RedirectWithFlash(c, bookURL(uint(id)), "copy_added", 0).
	RedirectWithFlash(c, bookURL(uint(id)), "copy_added", 0)
}

// SetCopyStatus handles POST /panel/copies/:id/status.
func (h *WebPanelLoansHandler) SetCopyStatus(c *gin.Context) {
	// 1. id, err := strconv.ParseUint(c.Param("id"), 10, 32). If err != nil or id == 0: NotFound; return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	// 2. var f copyStatusForm. If err := c.ShouldBind(&f); err != nil: RedirectWithFlash(c, "/catalog", "generic_error", 0); return.
	var f copyStatusForm
	if err := c.ShouldBind(&f); err != nil {
		RedirectWithFlash(c, "/catalog", "generic_error", 0)
		return
	}
	// 3. err = h.panel.UpdateCopyStatus(uint(id), f.Status).
	// 4. ErrCopyNotAvailable -> RedirectWithFlash(c, bookURL(f.BookID), "copy_status_blocked", 0); return. ErrNotFound -> NotFound; return. Other error -> ServerError; return.
	err = h.panel.UpdateCopyStatus(uint(id), f.Status)
	if err != nil {
		if errors.Is(err, repository.ErrCopyNotAvailable) {
			RedirectWithFlash(c, bookURL(f.BookID), "copy_status_blocked", 0)
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			h.render.NotFound(c)
			return
		}
		h.render.ServerError(c)
		return
	}
	// 5. RedirectWithFlash(c, bookURL(f.BookID), "copy_status_updated", 0).
	RedirectWithFlash(c, bookURL(f.BookID), "copy_status_updated", 0)
}
