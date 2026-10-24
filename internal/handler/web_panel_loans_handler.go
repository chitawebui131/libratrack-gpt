package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type WebPanelLoansHandler struct {
	panel  repository.PanelRepository
	views  repository.LoanViewRepository
	loans  LoanService
	users  repository.UserRepository
	fines  FineCalculator
	render *Renderer
}

func NewWebPanelLoansHandler(panel repository.PanelRepository, views repository.LoanViewRepository, loans LoanService, users repository.UserRepository, fines FineCalculator, render *Renderer) *WebPanelLoansHandler {
	return &WebPanelLoansHandler{panel: panel, views: views, loans: loans, users: users, fines: fines, render: render}
}

type PanelLoanRow struct {
	Loan    model.LoanView
	Overdue bool
	Fine    int
}

type panelLoansPage struct {
	Rows  []PanelLoanRow
	Pager Pager
	Total int64
}

// LoansPage handles GET /panel/loans: active loans, 20 per page.
func (h *WebPanelLoansHandler) LoansPage(c *gin.Context) {
	// 1. page := parsePage(c.Query("page")); const limit = 20.
	page := parsePage(c.Query("page"))
	const limit = 20

	// 2. views, total, err := h.views.ListActiveViews(page, limit). On error: h.render.ServerError(c); return.
	views, total, err := h.views.ListActiveViews(page, limit)
	if err != nil {
		h.render.ServerError(c)
		return
	}

	// 3. now := time.Now().UTC(); rows := []PanelLoanRow{}.
	now := time.Now().UTC()
	rows := []PanelLoanRow{}

	// 4. For every v in views append PanelLoanRow{Loan: v, Overdue: v.IsOverdue(now), Fine: h.fines.CalculateFine(v.DueAt, now)}.
	for _, v := range views {
		rows = append(rows, PanelLoanRow{Loan: v, Overdue: v.IsOverdue(now), Fine: h.fines.CalculateFine(v.DueAt, now)})
	}

	// 5. h.render.Render(c, http.StatusOK, "panel_loans", "loans", panelLoansPage{Rows: rows, Pager: BuildPager(c.Request, page, limit, total), Total: total}).
	h.render.Render(c, http.StatusOK, "panel_loans", "loans", panelLoansPage{Rows: rows, Pager: BuildPager(c.Request, page, limit, total), Total: total})
}
