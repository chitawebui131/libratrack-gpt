package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type WebPanelReportsHandler struct {
	reports repository.ReportRepository
	fines   FineCalculator
	render  *Renderer
}

func NewWebPanelReportsHandler(reports repository.ReportRepository, fines FineCalculator, render *Renderer) *WebPanelReportsHandler {
	return &WebPanelReportsHandler{reports: reports, fines: fines, render: render}
}

type OverdueRow struct {
	Loan model.LoanView
	Days int
	Fine int
}

type reportsPage struct {
	Top     []model.TopBook
	Overdue []OverdueRow
}

// utcDay returns midnight UTC of the date of t.
func utcDay(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

// overdueDays returns the whole calendar days from due to now (UTC dates). Never negative.
func overdueDays(due, now time.Time) int {
	days := int(utcDay(now).Sub(utcDay(due)) / (24 * time.Hour))
	if days < 0 {
		return 0
	}
	return days
}

// Reports handles GET /panel/reports.
func (h *WebPanelReportsHandler) Reports(c *gin.Context) {
	// 1. now := time.Now().UTC().
	now := time.Now().UTC()

	// 2. top, err := h.reports.TopBooks(10). On error: h.render.ServerError(c); return.
	top, err := h.reports.TopBooks(10)
	if err != nil {
		h.render.ServerError(c)
		return
	}

	// 3. loans, err := h.reports.OverdueLoans(). On error: ServerError; return.
	loans, err := h.reports.OverdueLoans(now)
	if err != nil {
		h.render.ServerError(c)
		return
	}

	// 4. rows := []OverdueRow{}. For every v in loans append OverdueRow{Loan: v, Days: overdueDays(v.DueAt, now), Fine: h.fines.CalculateFine(v.DueAt, now)}.
	rows := []OverdueRow{}
	for _, v := range loans {
		rows = append(rows, OverdueRow{Days: overdueDays(v.DueAt, now), Fine: h.fines.CalculateFine(v.DueAt, now)})
	}

	// 5. h.render.Render(c, http.StatusOK, "panel_reports", "reports", reportsPage{Top: top, Overdue: rows}).
	h.render.Render(c, http.StatusOK, "panel_reports", "reports", reportsPage{Top: top, Overdue: rows})
}
