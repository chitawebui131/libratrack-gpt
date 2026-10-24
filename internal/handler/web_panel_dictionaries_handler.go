package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type WebPanelDictionariesHandler struct {
	dict   repository.DictionaryRepository
	render *Renderer
}

func NewWebPanelDictionariesHandler(dict repository.DictionaryRepository, render *Renderer) *WebPanelDictionariesHandler {
	return &WebPanelDictionariesHandler{dict: dict, render: render}
}

type authorForm struct {
	FullName string `form:"full_name" binding:"required,max=255"`
	Bio      string `form:"bio" binding:"max=2000"`
}

// authorsPage is the data of the authors page. EditID == 0 means create mode.
type authorsPage struct {
	Authors  []model.Author
	EditID   uint
	FullName string
	Bio      string
	Errors   map[string]string
}

func (h *WebPanelDictionariesHandler) renderAuthors(c *gin.Context, status int, p authorsPage) {
	// 1. items, err := h.dict.ListAuthors(). On error: h.render.ServerError(c); return. p.Authors = items.
	items, err := h.dict.ListAuthors()
	if err != nil {
		h.render.ServerError(c)
		return
	}
	p.Authors = items
	// 2. h.render.Render(c, status, "panel_authors", "authors", p).
	h.render.Render(c, status, "panel_authors", "authors", p)
}

// AuthorsPage handles GET /panel/authors.
func (h *WebPanelDictionariesHandler) AuthorsPage(c *gin.Context) {
	// h.renderAuthors(c, http.StatusOK, authorsPage{Errors: map[string]string{}}).
	h.renderAuthors(c, http.StatusOK, authorsPage{Errors: map[string]string{}})
}

// CreateAuthor handles POST /panel/authors.
func (h *WebPanelDictionariesHandler) CreateAuthor(c *gin.Context) {
	// 1. var f authorForm. If err := c.ShouldBind(&f); err != nil: h.renderAuthors(c, 422, authorsPage{FullName: f.FullName, Bio: f.Bio, Errors: fieldErrors(err)}); return.
	var f authorForm
	if err := c.ShouldBind(&f); err != nil {
		h.renderAuthors(c, 422, authorsPage{FullName: f.FullName, Bio: f.Bio, Errors: fieldErrors(err)})
		return
	}
	// 2. If err := h.dict.CreateAuthor(&model.Author{FullName: f.FullName, Bio: f.Bio}); err != nil: h.render.ServerError(c); return.
	if err := h.dict.CreateAuthor(&model.Author{FullName: f.FullName, Bio: f.Bio}); err != nil {
		h.render.ServerError(c)
		return
	}
	// 3. RedirectWithFlash(c, "/panel/authors", "author_saved", 0).
	RedirectWithFlash(c, "/panel/authors", "author_saved", 0)
}

// EditAuthorPage handles GET /panel/authors/:id/edit.
func (h *WebPanelDictionariesHandler) EditAuthorPage(c *gin.Context) {
	// 1. id, err := strconv.ParseUint(c.Param("id"), 10, 32). If err != nil or id == 0: NotFound; return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	// 2. a, err := h.dict.GetAuthor(uint(id)). If errors.Is(err, repository.ErrNotFound): NotFound; return. Other error: ServerError; return.
	a, err := h.dict.GetAuthor(uint(id))
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 3. h.renderAuthors(c, http.StatusOK, authorsPage{EditID: a.ID, FullName: a.FullName, Bio: a.Bio, Errors: map[string]string{}}).
	h.renderAuthors(c, http.StatusOK, authorsPage{EditID: a.ID, FullName: a.FullName, Bio: a.Bio, Errors: map[string]string{}})
}

// UpdateAuthor handles POST /panel/authors/:id.
func (h *WebPanelDictionariesHandler) UpdateAuthor(c *gin.Context) {
	// 1. Parse the id like in EditAuthorPage (NotFound on error).
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	// 2. var f authorForm. On bind error: h.renderAuthors(c, 422, authorsPage{EditID: uint(id), FullName: f.FullName, Bio: f.Bio, Errors: fieldErrors(err)}); return.
	var f authorForm
	if err := c.ShouldBind(&f); err != nil {
		h.renderAuthors(c, 422, authorsPage{EditID: uint(id), FullName: f.FullName, Bio: f.Bio, Errors: fieldErrors(err)})
		return
	}
	// 3. err := h.dict.UpdateAuthor(&model.Author{ID: uint(id), FullName: f.FullName, Bio: f.Bio}). ErrNotFound -> NotFound; other error -> ServerError.
	err = h.dict.UpdateAuthor(&model.Author{ID: uint(id), FullName: f.FullName, Bio: f.Bio})
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 4. RedirectWithFlash(c, "/panel/authors", "author_saved", 0).
	RedirectWithFlash(c, "/panel/authors", "author_saved", 0)
}
