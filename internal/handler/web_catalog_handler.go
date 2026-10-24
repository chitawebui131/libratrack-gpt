package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type WebCatalogHandler struct {
	catalog repository.CatalogRepository
	dict    repository.DictionaryRepository
	render  *Renderer
}

func NewWebCatalogHandler(catalog repository.CatalogRepository, dict repository.DictionaryRepository, render *Renderer) *WebCatalogHandler {
	return &WebCatalogHandler{catalog: catalog, dict: dict, render: render}
}

type BookCard struct {
	Book      model.Book
	Total     int
	Available int
}

type catalogPage struct {
	Cards         []BookCard
	Categories    []model.Category
	Authors       []model.Author
	Query         string
	CategoryID    uint
	AuthorID      uint
	OnlyAvailable bool
	Pager         Pager
	Total         int64
}

// Catalog handles GET /catalog (12 books per page).
func (h *WebCatalogHandler) Catalog(c *gin.Context) {
	// 1. q := strings.TrimSpace(c.Query("q")); categoryID := parseUintParam(c.Query("category_id")); authorID := parseUintParam(c.Query("author_id")); only := c.Query("available") == "1"; page := parsePage(c.Query("page")); const limit = 12.
	q := strings.TrimSpace(c.Query("q"))
	categoryID := parseUintParam(c.Query("category_id"))
	authorID := parseUintParam(c.Query("author_id"))
	only := c.Query("available") == "1"
	page := parsePage(c.Query("page"))
	const limit = 12

	// 2. books, total, err := h.catalog.Search(repository.BookFilter{Query: q, CategoryID: categoryID, AuthorID: authorID, OnlyAvailable: only, Page: page, Limit: limit}). On error: h.render.ServerError(c); return.
	books, total, err := h.catalog.Search(repository.BookFilter{Query: q, CategoryID: categoryID, AuthorID: authorID, OnlyAvailable: only, Page: page, Limit: limit})
	if err != nil {
		h.render.ServerError(c)
		return
	}

	// 3. ids := the IDs of books. avail, err := h.catalog.AvailabilityForBooks(ids) (ONE call). On error ServerError.
	ids := make([]uint, len(books))
	for i, book := range books {
		ids[i] = book.ID
	}
	avail, err := h.catalog.AvailabilityForBooks(ids)
	if err != nil {
		h.render.ServerError(c)
		return
	}

	// 4. cards := []BookCard{}. For every book: a := avail[book.ID]; append BookCard{Book: book, Total: a.Total, Available: a.Available}.
	cards := make([]BookCard, len(books))
	for i, book := range books {
		a := avail[book.ID]
		cards[i] = BookCard{Book: book, Total: a.Total, Available: a.Available}
	}

	// 5. categories, err := h.dict.ListCategories(); authors, err := h.dict.ListAuthors(). On error ServerError.
	categories, err := h.dict.ListCategories()
	if err != nil {
		h.render.ServerError(c)
		return
	}
	authors, err := h.dict.ListAuthors()
	if err != nil {
		h.render.ServerError(c)
		return
	}

	// 6. h.render.Render(c, http.StatusOK, "catalog", "catalog", catalogPage{Cards: cards, Categories: categories, Authors: authors, Query: q, CategoryID: categoryID, AuthorID: authorID, OnlyAvailable: only, Pager: BuildPager(c.Request, page, limit, total), Total: total}).
	h.render.Render(c, http.StatusOK, "catalog", "catalog", catalogPage{Cards: cards, Categories: categories, Authors: authors, Query: q, CategoryID: categoryID, AuthorID: authorID, OnlyAvailable: only, Pager: BuildPager(c.Request, page, limit, total), Total: total})
}
