package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
)

// ReturnLoan handles POST /panel/loans/:id/return.
func (h *WebPanelLoansHandler) ReturnLoan(c *gin.Context) {
	// 1. id, err := strconv.ParseUint(c.Param("id"), 10, 32). If err != nil or id == 0: h.render.NotFound(c); return.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.render.NotFound(c)
		return
	}

	// 2. loan, err := h.loans.Return(uint(id)).
	loan, err := h.loans.Return(uint(id))

	// 3. ErrNotFound -> h.render.NotFound(c); return. ErrAlreadyReturned -> RedirectWithFlash(c, "/panel/loans", "generic_error", 0); return. Other error -> h.render.ServerError(c); return.
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.render.NotFound(c)
			return
		case errors.Is(err, repository.ErrAlreadyReturned):
			RedirectWithFlash(c, "/panel/loans", "generic_error", 0)
			return
		default:
			h.render.ServerError(c)
			return
		}
	}

	// 4. RedirectWithFlash(c, "/panel/loans", "loan_returned", int64(loan.FineAmount)).
	RedirectWithFlash(c, "/panel/loans", "loan_returned", int64(loan.FineAmount))
}
