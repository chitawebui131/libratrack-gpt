package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
)

// List handles GET /loans (librarian and admin).
// List godoc
// @Summary      List loans
// @Description  Get a list of all book loans
// @Tags         loans
// @Accept       json
// @Produce      json
// @Success      200  {array}   model.Loan
// @Security     BearerAuth
// @Router       /loans [get]
func (h *LoanHandler) List(c *gin.Context) {
	// 1. status := c.Query("status"). If status != "" and !model.IsValidLoanStatus(status): respond 422 and return.
	status := c.Query("status")
	if status != "" && !model.IsValidLoanStatus(status) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid status"})
		return
	}

	// 2. page, _ := strconv.Atoi(c.Query("page")) and limit, _ := strconv.Atoi(c.Query("limit")). A value of 0 means "use the default".
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	// 3. loans, err := h.svc.List(status, page, limit). If err != nil: respond 500 and return.
	loans, err := h.svc.List(status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	// 4. c.JSON(http.StatusOK, loans).
	c.JSON(http.StatusOK, loans)
}

// ListMine handles GET /me/loans. The user id comes ONLY from the token.
func (h *LoanHandler) ListMine(c *gin.Context) {
	// 1. userID, _, ok := middleware.CurrentUser(c). If !ok: respond 401 and return.
	userID, _, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	// 2. Read page and limit like in List.
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	// 3. loans, err := h.svc.ListByUser(userID, page, limit). If err != nil: respond 500 and return.
	loans, err := h.svc.ListByUser(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	// 4. c.JSON(http.StatusOK, loans).
	c.JSON(http.StatusOK, loans)
}
