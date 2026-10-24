package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type categoryForm struct {
	Name string `form:"name" binding:"required,max=100"`
}

// categoriesPage is the data of the categories page. EditID == 0 means create mode.
type categoriesPage struct {
	Categories []model.Category
	EditID     uint
	Name       string
	Errors     map[string]string
}

func (h *WebPanelDictionariesHandler) renderCategories(c *gin.Context, status int, p categoriesPage) {
	// 1. items, err := h.dict.ListCategories(). On error: h.render.ServerError(c); return. p.Categories = items.
	items, err := h.dict.ListCategories()
	if err != nil {
		h.render.ServerError(c)
		return
	}
	p.Categories = items
	// 2. h.render.Render(c, status, "panel_categories", "categories", p).
	h.render.Render(c, status, "panel_categories", "categories", p)
}

// CategoriesPage handles GET /panel/categories.
func (h *WebPanelDictionariesHandler) CategoriesPage(c *gin.Context) {
	// h.renderCategories(c, http.StatusOK, categoriesPage{Errors: map[string]string{}}).
	h.renderCategories(c, http.StatusOK, categoriesPage{Errors: map[string]string{}})
}

// CreateCategory handles POST /panel/categories.
func (h *WebPanelDictionariesHandler) CreateCategory(c *gin.Context) {
	// 1. var f categoryForm. On bind error: h.renderCategories(c, 422, categoriesPage{Name: f.Name, Errors: fieldErrors(err)}); return.
	var f categoryForm
	if err := c.ShouldBind(&f); err != nil {
		h.renderCategories(c, 422, categoriesPage{Name: f.Name, Errors: fieldErrors(err)})
		return
	}
	// 2. err := h.dict.CreateCategory(&model.Category{Name: f.Name}).
	err := h.dict.CreateCategory(&model.Category{Name: f.Name})
	// 3. If errors.Is(err, repository.ErrConflict): h.renderCategories(c, 409, categoriesPage{Name: f.Name, Errors: map[string]string{"Name": "name_taken"}}); return.
	if errors.Is(err, repository.ErrConflict) {
		h.renderCategories(c, 409, categoriesPage{Name: f.Name, Errors: map[string]string{"Name": "name_taken"}})
		return
	}
	// 4. If err != nil: h.render.ServerError(c); return.
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 5. RedirectWithFlash(c, "/panel/categories", "category_saved", 0).
	RedirectWithFlash(c, "/panel/categories", "category_saved", 0)
}

// EditCategoryPage handles GET /panel/categories/:id/edit.
func (h *WebPanelDictionariesHandler) EditCategoryPage(c *gin.Context) {
	// 1. Parse the id (NotFound on error or 0). cat, err := h.dict.GetCategory(uint(id)). ErrNotFound -> NotFound; other error -> ServerError.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	cat, err := h.dict.GetCategory(uint(id))
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 2. h.renderCategories(c, http.StatusOK, categoriesPage{EditID: cat.ID, Name: cat.Name, Errors: map[string]string{}}).
	h.renderCategories(c, http.StatusOK, categoriesPage{EditID: cat.ID, Name: cat.Name, Errors: map[string]string{}})
}

// UpdateCategory handles POST /panel/categories/:id.
func (h *WebPanelDictionariesHandler) UpdateCategory(c *gin.Context) {
	// 1. Parse the id (NotFound on error or 0).
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}
	// 2. var f categoryForm. On bind error: render 422 with EditID, Name and fieldErrors(err); return.
	var f categoryForm
	if err := c.ShouldBind(&f); err != nil {
		h.renderCategories(c, 422, categoriesPage{EditID: uint(id), Name: f.Name, Errors: fieldErrors(err)})
		return
	}
	// 3. err := h.dict.UpdateCategory(&model.Category{ID: uint(id), Name: f.Name}).
	//    ErrConflict -> render 409 with EditID, Name and Errors{"Name": "name_taken"}. ErrNotFound -> NotFound. Other error -> ServerError.
	err = h.dict.UpdateCategory(&model.Category{ID: uint(id), Name: f.Name})
	if errors.Is(err, repository.ErrConflict) {
		h.renderCategories(c, 409, categoriesPage{EditID: uint(id), Name: f.Name, Errors: map[string]string{"Name": "name_taken"}})
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		h.render.NotFound(c)
		return
	}
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 4. RedirectWithFlash(c, "/panel/categories", "category_saved", 0).
	RedirectWithFlash(c, "/panel/categories", "category_saved", 0)
}
