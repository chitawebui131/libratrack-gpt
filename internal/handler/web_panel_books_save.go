package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
	"libratrack/internal/service"
)

// saveBook is shared by create (id == 0) and update.
func (h *WebPanelBooksHandler) saveBook(c *gin.Context, id uint) {
	var f bookForm
	if err := c.ShouldBind(&f); err != nil {
		h.renderForm(c, http.StatusUnprocessableEntity, newBookFormPage(id, f, fieldErrors(err)))
		return
	}

	// 1. isbn, err := service.NormalizeISBN(f.ISBN). If err != nil: h.renderForm(c, http.StatusUnprocessableEntity, newBookFormPage(id, f, map[string]string{"ISBN": "isbn_invalid"})); return.
	isbn, err := service.NormalizeISBN(f.ISBN)
	if err != nil {
		h.renderForm(c, http.StatusUnprocessableEntity, newBookFormPage(id, f, map[string]string{"ISBN": "isbn_invalid"}))
		return
	}

	// 2. lang, ok := service.NormalizeLanguage(f.Language). If !ok: h.renderForm(c, http.StatusUnprocessableEntity, newBookFormPage(id, f, map[string]string{"Language": "language_invalid"})); return.
	lang, ok := service.NormalizeLanguage(f.Language)
	if !ok {
		h.renderForm(c, http.StatusUnprocessableEntity, newBookFormPage(id, f, map[string]string{"Language": "language_invalid"}))
		return
	}

	// 3. Build the model.Book with ISBN: isbn, UDK: strings.TrimSpace(f.UDK), Language: lang (in addition to the fields that are already set).
	b := &model.Book{ID: id, Title: f.Title, ISBN: isbn, CategoryID: f.CategoryID, Publisher: f.Publisher, PublishedYear: f.PublishedYear, Description: f.Description, UDK: strings.TrimSpace(f.UDK), Language: lang, PublicationCity: strings.TrimSpace(f.PublicationCity), PageCount: f.PageCount, BBK: strings.TrimSpace(f.BBK)}
	err = h.panel.SaveBook(b, f.AuthorIDs)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			h.renderForm(c, http.StatusConflict, newBookFormPage(id, f, map[string]string{"ISBN": "isbn_taken"}))
			return
		} else if errors.Is(err, repository.ErrInvalidReference) {
			h.renderForm(c, http.StatusUnprocessableEntity, newBookFormPage(id, f, map[string]string{"CategoryID": "invalid_reference"}))
			return
		} else if errors.Is(err, repository.ErrNotFound) {
			h.render.NotFound(c)
			return
		} else {
			h.render.ServerError(c)
			return
		}
	}
	code := "book_created"
	if id != 0 {
		code = "book_updated"
	}
	RedirectWithFlash(c, "/books/"+strconv.FormatUint(uint64(b.ID), 10), code, 0)
}

// CreateBook handles POST /panel/books.
func (h *WebPanelBooksHandler) CreateBook(c *gin.Context) {
	h.saveBook(c, 0)
}

// UpdateBook handles POST /panel/books/:id.
func (h *WebPanelBooksHandler) UpdateBook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	h.saveBook(c, uint(id))
}
