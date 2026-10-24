package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/repository"
)

// FineCalculator is implemented by service.LoanPolicy.
type FineCalculator interface {
	CalculateFine(dueAt, returnedAt time.Time) int
}

type WebLoansHandler struct {
	views  repository.LoanViewRepository
	fines  FineCalculator
	render *Renderer
}

func NewWebLoansHandler(views repository.LoanViewRepository, fines FineCalculator, render *Renderer) *WebLoansHandler {
	return &WebLoansHandler{views: views, fines: fines, render: render}
}

type MyLoanRow struct {
	Loan    model.LoanView
	Overdue bool
	Fine    int
}

type myLoansPage struct {
	Rows  []MyLoanRow
	Pager Pager
	Total int64
}

// MyLoans handles GET /my/loans.
func (h *WebLoansHandler) MyLoans(c *gin.Context) {
	// 1. user, ok := middleware.WebUser(c). If !ok: c.Redirect(http.StatusSeeOther, "/login"); return.
	user, ok := middleware.WebUser(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// 2. page := parsePage(c.Query("page")); const limit = 20.
	page := parsePage(c.Query("page"))
	const limit = 20

	// 3. views, total, err := h.views.ListViewsByUser(user.ID, page, limit). On error: h.render.ServerError(c); return.
	views, total, err := h.views.ListViewsByUser(user.ID, page, limit)
	if err != nil {
		h.render.ServerError(c)
		return
	}

	// 4. now := time.Now().UTC(); rows := []MyLoanRow{}.
	now := time.Now().UTC()
	rows := []MyLoanRow{}

	// 5. For every v in views: fine := v.FineAmount; if v.IsActive() { fine = h.fines.CalculateFine(v.DueAt, now) }. Append MyLoanRow{Loan: v, Overdue: v.IsOverdue(now), Fine: fine}.
	for _, v := range views {
		fine := v.FineAmount
		if v.IsActive() {
			fine = h.fines.CalculateFine(v.DueAt, now)
		}
		rows = append(rows, MyLoanRow{Loan: v, Overdue: v.IsOverdue(now), Fine: fine})
	}

	// 6. h.render.Render(c, http.StatusOK, "my_loans", "my_loans", myLoansPage{Rows: rows, Pager: BuildPager(c.Request, page, limit, total), Total: total}).
	h.render.Render(c, http.StatusOK, "my_loans", "my_loans", myLoansPage{Rows: rows, Pager: BuildPager(c.Request, page, limit, total), Total: total})
}
