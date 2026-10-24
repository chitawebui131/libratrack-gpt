package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/service"
)

// ImportRunner is implemented by *service.ImportService.
type ImportRunner interface {
	Run(rows []service.RawRow, dryRun bool) *model.ImportReport
}

type WebPanelImportHandler struct {
	importer ImportRunner
	export   ExportSource
	render   *Renderer
}

func NewWebPanelImportHandler(importer ImportRunner, export ExportSource, render *Renderer) *WebPanelImportHandler {
	return &WebPanelImportHandler{importer: importer, export: export, render: render}
}

// importPage is the data of the import page. Error is an ASCII key shown by the template ("", "no_file", "bad_type", "bad_file").
type importPage struct {
	Report *model.ImportReport
	Error  string
}

// ImportPage handles GET /panel/import.
func (h *WebPanelImportHandler) ImportPage(c *gin.Context) {
	h.render.Render(c, http.StatusOK, "panel_import", "import", importPage{})
}

// ImportUpload handles POST /panel/import (multipart form: file, dry_run).
func (h *WebPanelImportHandler) ImportUpload(c *gin.Context) {
	// 1. fh, err := c.FormFile("file"). If err != nil: render page "panel_import" with status 400 and importPage{Error: "no_file"}; return.
	fh, err := c.FormFile("file")
	if err != nil {
		h.render.Render(c, http.StatusBadRequest, "panel_import", "import", importPage{Error: "no_file"})
		return
	}
	// 2. ext := strings.ToLower(filepath.Ext(fh.Filename)). If ext != ".csv" and ext != ".xlsx": render with status 400 and importPage{Error: "bad_type"}; return.
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext != ".csv" && ext != ".xlsx" {
		h.render.Render(c, http.StatusBadRequest, "panel_import", "import", importPage{Error: "bad_type"})
		return
	}
	// 3. f, err := fh.Open(). If err != nil: h.render.ServerError(c); return. defer f.Close().
	f, err := fh.Open()
	if err != nil {
		h.render.ServerError(c)
		return
	}
	defer f.Close()
	// 4. var rows []service.RawRow. If ext == ".csv" { rows, err = service.ReadCSV(f) } else { rows, err = service.ReadXLSX(f) }.
	var rows []service.RawRow
	if ext == ".csv" {
		rows, err = service.ReadCSV(f)
	} else {
		rows, err = service.ReadXLSX(f)
	}
	// 5. If err != nil: render page with status 422 and importPage{Error: "bad_file"}; return.
	if err != nil {
		h.render.Render(c, http.StatusUnprocessableEntity, "panel_import", "import", importPage{Error: "bad_file"})
		return
	}
	// 6. dryRun := c.PostForm("dry_run") == "1".
	dryRun := c.PostForm("dry_run") == "1"
	// 7. report := h.importer.Run(rows, dryRun).
	report := h.importer.Run(rows, dryRun)
	// 8. h.render.Render(c, http.StatusOK, "panel_import", "import", importPage{Report: report}).
	service.TrimDetails(report, 300)
	h.render.Render(c, http.StatusOK, "panel_import", "import", importPage{Report: report})
}
