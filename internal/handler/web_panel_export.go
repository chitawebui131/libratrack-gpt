package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/service"
)

// ExportSource is implemented by repository.ExportRepository.
type ExportSource interface {
	ListExportRows() ([]model.ExportRow, error)
}

// ExportFile handles GET /panel/export?format=csv|xlsx.
func (h *WebPanelImportHandler) ExportFile(c *gin.Context) {
	// 1. format := c.DefaultQuery("format", "csv"). If format != "csv" and format != "xlsx": h.render.NotFound(c); return.
	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "xlsx" {
		h.render.NotFound(c)
		return
	}
	// 2. rows, err := h.export.ListExportRows(). If err != nil: h.render.ServerError(c); return.
	rows, err := h.export.ListExportRows()
	if err != nil {
		h.render.ServerError(c)
		return
	}
	// 3. name := "books-" + time.Now().UTC().Format("20060102") + "." + format.
	name := "books-" + time.Now().UTC().Format("20060102") + "." + format
	// 4. c.Header("Content-Disposition", "attachment; filename=\""+name+"\"").
	c.Header("Content-Disposition", "attachment; filename=\""+name+"\"")
	// 5. If format == "csv": c.Header("Content-Type", "text/csv; charset=utf-8"); c.Status(http.StatusOK); err = service.WriteCSV(c.Writer, rows).
	//    Otherwise: c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"); c.Status(http.StatusOK); err = service.WriteXLSX(c.Writer, rows).
	if format == "csv" {
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Status(http.StatusOK)
		err = service.WriteCSV(c.Writer, rows)
	} else {
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Status(http.StatusOK)
		err = service.WriteXLSX(c.Writer, rows)
	}
	// 6. If err != nil: c.Error(err) (the headers are already sent, so nothing else can be written).
	if err != nil {
		c.Error(err)
	}
}
