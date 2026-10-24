package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

// panelBookRow is one table row of the staff books list.
type panelBookRow struct {
	Book      model.Book
	Total     int
	Available int
}

type panelBooksPage struct {
	Rows  []panelBookRow
	Query string
	Pager Pager
}

// BooksPage handles GET /panel/books (20 books per page, optional ?q= search).
func (h *WebPanelBooksHandler) BooksPage(c *gin.Context) {
	// 1. q := strings.TrimSpace(c.Query("q")); page := parsePage(c.Query("page")); const limit = 20.
	q := strings.TrimSpace(c.Query("q"))
	page := parsePage(c.Query("page"))
	const limit = 20

	// 2. books, total, err := h.catalog.Search(repository.BookFilter{Query: q, Page: page, Limit: limit}). On error: h.render.ServerError(c); return.
	books, total, err := h.catalog.Search(repository.BookFilter{Query: q, Page: page, Limit: limit})
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

	// 4. rows := one panelBookRow per book with Total and Available from avail[book.ID].
	rows := make([]panelBookRow, len(books))
	for i, book := range books {
		rows[i] = panelBookRow{
			Book:      book,
			Total:     avail[book.ID].Total,
			Available: avail[book.ID].Available,
		}
	}

	// 5. h.render.Render(c, http.StatusOK, "panel_books", "books", panelBooksPage{Rows: rows, Query: q, Pager: BuildPager(c.Request, page, limit, total)}).
	h.render.Render(c, http.StatusOK, "panel_books", "books", panelBooksPage{Rows: rows, Query: q, Pager: BuildPager(c.Request, page, limit, total)})
}
