package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"libratrack/internal/repository"
)

type WebPanelDashboardHandler struct {
	reports repository.ReportRepository
	render  *Renderer
}

func NewWebPanelDashboardHandler(reports repository.ReportRepository, render *Renderer) *WebPanelDashboardHandler {
	return &WebPanelDashboardHandler{reports: reports, render: render}
}

// Dashboard handles GET /panel.
func (h *WebPanelDashboardHandler) Dashboard(c *gin.Context) {
	// 1. stats, err := h.reports.Stats(time.Now().UTC()). On error: h.render.ServerError(c); return.
	stats, err := h.reports.Stats(time.Now().UTC())
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 2. h.render.Render(c, http.StatusOK, "panel_dashboard", "panel", stats).
	h.render.Render(c, http.StatusOK, "panel_dashboard", "panel", stats)
}
