package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type WebPanelBooksHandler struct {
	panel   repository.PanelRepository
	catalog repository.CatalogRepository
	dict    repository.DictionaryRepository
	render  *Renderer
}

func NewWebPanelBooksHandler(panel repository.PanelRepository, catalog repository.CatalogRepository, dict repository.DictionaryRepository, render *Renderer) *WebPanelBooksHandler {
	return &WebPanelBooksHandler{panel: panel, catalog: catalog, dict: dict, render: render}
}

type bookForm struct {
	Title           string `form:"title" binding:"required,max=255"`
	ISBN            string `form:"isbn" binding:"omitempty,max=30"`
	CategoryID      uint   `form:"category_id" binding:"required,gt=0"`
	AuthorIDs       []uint `form:"author_ids" binding:"required,min=1,dive,gt=0"`
	Publisher       string `form:"publisher" binding:"max=255"`
	PublishedYear   int    `form:"published_year" binding:"omitempty,gte=1000,lte=2100"`
	Description     string `form:"description" binding:"max=5000"`
	UDK             string `form:"udk" binding:"max=100"`
	Language        string `form:"language" binding:"max=20"`
	PublicationCity string `form:"publication_city" binding:"max=100"`
	PageCount       int    `form:"page_count" binding:"omitempty,gte=0,lte=100000"`
	BBK             string `form:"bbk" binding:"max=30"`
}

// bookFormPage is the data of the create/edit page. ID == 0 means create mode.
type bookFormPage struct {
	ID              uint
	Title           string
	ISBN            string
	CategoryID      uint
	Publisher       string
	PublishedYear   int
	Description     string
	UDK             string
	Language        string
	PublicationCity string
	PageCount       int
	BBK             string
	AuthorSelected  map[uint]bool
	Categories      []model.Category
	Authors         []model.Author
	Errors          map[string]string
}

// newBookFormPage builds the page data from submitted form values.
func newBookFormPage(id uint, f bookForm, errs map[string]string) bookFormPage {
	sel := map[uint]bool{}
	for _, a := range f.AuthorIDs {
		sel[a] = true
	}
	return bookFormPage{ID: id, Title: f.Title, ISBN: f.ISBN, CategoryID: f.CategoryID, Publisher: f.Publisher, PublishedYear: f.PublishedYear, Description: f.Description, UDK: f.UDK, Language: f.Language, PublicationCity: f.PublicationCity, PageCount: f.PageCount, BBK: f.BBK, AuthorSelected: sel, Errors: errs}
}

// fillLists loads the categories and authors for the select boxes.
func (h *WebPanelBooksHandler) fillLists(p *bookFormPage) error {
	cats, err := h.dict.ListCategories()
	if err != nil {
		return err
	}
	p.Categories = cats
	authors, err := h.dict.ListAuthors()
	if err != nil {
		return err
	}
	p.Authors = authors
	return nil
}

// renderForm fills the lists and renders the form page with the given status.
func (h *WebPanelBooksHandler) renderForm(c *gin.Context, status int, p bookFormPage) {
	if err := h.fillLists(&p); err != nil {
		h.render.ServerError(c)
		return
	}
	title := "book_new"
	if p.ID != 0 {
		title = "book_edit"
	}
	h.render.Render(c, status, "panel_book_form", title, p)
}

// NewBookPage handles GET /panel/new-book.
func (h *WebPanelBooksHandler) NewBookPage(c *gin.Context) {
	h.renderForm(c, http.StatusOK, newBookFormPage(0, bookForm{}, map[string]string{}))
}

// EditBookPage handles GET /panel/books/:id/edit.
func (h *WebPanelBooksHandler) EditBookPage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	book, err := h.catalog.GetBook(uint(id))
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	} else if err != nil {
		h.render.ServerError(c)
		return
	}
	f := bookForm{Title: book.Title, ISBN: book.ISBN, CategoryID: book.CategoryID, Publisher: book.Publisher, PublishedYear: book.PublishedYear, Description: book.Description, UDK: book.UDK, Language: book.Language, PublicationCity: book.PublicationCity, PageCount: book.PageCount, BBK: book.BBK}
	for _, author := range book.Authors {
		f.AuthorIDs = append(f.AuthorIDs, author.ID)
	}
	h.renderForm(c, http.StatusOK, newBookFormPage(uint(id), f, map[string]string{}))
}
