package handler

import (
	"net/http"

	"libratrack/internal/model"
	"libratrack/internal/repository"

	"github.com/gin-gonic/gin"
)

type BookHandler struct {
	repo repository.BookRepository
}

func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

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
