package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
)

type issueForm struct {
	CopyID      uint   `form:"copy_id" binding:"required,gt=0"`
	ReaderEmail string `form:"reader_email" binding:"required,email"`
	BookID      uint   `form:"book_id" binding:"required,gt=0"`
}

// IssueLoan handles POST /panel/loans.
func (h *WebPanelLoansHandler) IssueLoan(c *gin.Context) {
	// 1. var f issueForm. If err := c.ShouldBind(&f); err != nil: target := "/panel/loans"; if f.BookID != 0 { target = bookURL(f.BookID) }; RedirectWithFlash(c, target, "generic_error", 0); return.
	var f issueForm
	if err := c.ShouldBind(&f); err != nil {
		target := "/panel/loans"
		if f.BookID != 0 {
			target = bookURL(f.BookID)
		}
		RedirectWithFlash(c, target, "generic_error", 0)
		return
	}
	// 2. target := bookURL(f.BookID).
	target := bookURL(f.BookID)
	// 3. user, err := h.users.FindByEmail(strings.ToLower(strings.TrimSpace(f.ReaderEmail))).
	//    If errors.Is(err, repository.ErrNotFound): RedirectWithFlash(c, target, "loan_error_reader_not_found", 0); return. Other error: ServerError; return.
	user, err := h.users.FindByEmail(strings.ToLower(strings.TrimSpace(f.ReaderEmail)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RedirectWithFlash(c, target, "loan_error_reader_not_found", 0)
			return
		}
		h.render.ServerError(c)
		return
	}
	// 4. _, err = h.loans.Issue(f.CopyID, user.ID).
	// 5. Map errors with errors.Is and return after handling:
	//      repository.ErrCopyNotAvailable -> RedirectWithFlash(c, target, "loan_error_not_available", 0)
	//      repository.ErrLoanLimitReached -> RedirectWithFlash(c, target, "loan_error_limit_reached", 0)
	//      repository.ErrInvalidReference -> h.render.NotFound(c)
	//      any other error                -> h.render.ServerError(c)
	_, err = h.loans.Issue(f.CopyID, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrCopyNotAvailable) {
			RedirectWithFlash(c, target, "loan_error_not_available", 0)
			return
		}
		if errors.Is(err, repository.ErrLoanLimitReached) {
			RedirectWithFlash(c, target, "loan_error_limit_reached", 0)
			return
		}
		if errors.Is(err, repository.ErrInvalidReference) {
			h.render.NotFound(c)
			return
		}
		h.render.ServerError(c)
		return
	}
	// 6. RedirectWithFlash(c, target, "loan_issued", 0).
	RedirectWithFlash(c, target, "loan_issued", 0)
}
