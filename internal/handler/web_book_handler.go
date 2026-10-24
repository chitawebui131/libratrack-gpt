package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type bookPage struct {
	Book            model.Book
	Total           int
	Available       int
	IsStaff         bool
	Copies          []model.Copy
	AvailableCopies []model.Copy
}

// Book handles GET /books/:id.
func (h *WebCatalogHandler) Book(c *gin.Context) {
	// 1. id, err := strconv.ParseUint(c.Param("id"), 10, 32). If err != nil or id == 0: h.render.NotFound(c); return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}

	// 2. book, err := h.catalog.GetBook(uint(id)). If errors.Is(err, repository.ErrNotFound): NotFound; return. Any other error: ServerError; return.
	book, err := h.catalog.GetBook(uint(id))
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}

	// 3. user, _ := middleware.WebUser(c). isStaff := user != nil && (user.Role == model.RoleLibrarian || user.Role == model.RoleAdmin).
	user, _ := middleware.WebUser(c)
	isStaff := user != nil && (user.Role == model.RoleLibrarian || user.Role == model.RoleAdmin)

	// 4. page := bookPage{Book: *book, Total: len(book.Copies), IsStaff: isStaff}.
	page := bookPage{Book: *book, Total: len(book.Copies), IsStaff: isStaff}

	// 5. availableList := []model.Copy{}. For every copy cp in book.Copies: if cp.Status == model.CopyStatusAvailable, page.Available++ and append cp to availableList.
	availableList := []model.Copy{}
	for _, cp := range book.Copies {
		if cp.Status == model.CopyStatusAvailable {
			page.Available++
			availableList = append(availableList, cp)
		}
	}

	// 6. If isStaff: page.Copies = book.Copies and page.AvailableCopies = availableList. Otherwise: page.Book.Copies = nil.
	if isStaff {
		page.Copies = book.Copies
		page.AvailableCopies = availableList
	} else {
		page.Book.Copies = nil
	}

	// 7. h.render.Render(c, http.StatusOK, "book", book.Title, page).
	h.render.Render(c, http.StatusOK, "book", book.Title, page)
}
