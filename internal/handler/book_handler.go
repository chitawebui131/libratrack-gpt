package handler

import (
	"net/http"

	"libratrack/internal/model"
	"libratrack/internal/repository"

	"github.com/gin-gonic/gin"
)

// BookHandler handles book-related HTTP requests.
type BookHandler struct {
	repo repository.BookRepository
}

// NewBookHandler creates a new BookHandler.
func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

// @Summary Create a new book
// @Tags books
// @Accept json
// @Produce json
// @Param request body model.CreateBookRequest true "Book request"
// @Success 201 {object} model.Book
// @Failure 422 {object} gin.H
// @Failure 500 {object} gin.H
// @Security BearerAuth
// @Router /books [post]
// Create a new book.
func (h *BookHandler) Create(c *gin.Context) {
	var req model.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidation(c, err)
		return
	}

	book := model.Book{
		Title:         req.Title,
		ISBN:          req.ISBN,
		CategoryID:    req.CategoryID,
		Publisher:     req.Publisher,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	}

	if err := h.repo.Create(&book, req.AuthorIDs); err != nil {
		respondRepoError(c, err)
		return
	}

	c.JSON(http.StatusCreated, book)
}

// @Summary Get a book by ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path uint true "Book ID"
// @Success 200 {object} model.Book
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Security BearerAuth
// @Router /books/{id} [get]
// Get a book by ID.
func (h *BookHandler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	book, err := h.repo.FindByID(id)
	if err != nil {
		respondRepoError(c, err)
		return
	}

	c.JSON(http.StatusOK, book)
}
