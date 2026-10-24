package main

import (
	"gorm.io/gorm"

	"libratrack/internal/handler"
	"libratrack/internal/middleware"
	"libratrack/internal/repository"
	"libratrack/internal/service"
	"libratrack/web"
)

type webHandlers struct {
	Render         *handler.Renderer
	Cookies        middleware.CookieConfig
	Sec            securityHandlers
	Auth           *handler.WebAuthHandler
	Catalog        *handler.WebCatalogHandler
	Loans          *handler.WebLoansHandler
	PanelDashboard *handler.WebPanelDashboardHandler
	PanelBooks     *handler.WebPanelBooksHandler
	PanelDicts     *handler.WebPanelDictionariesHandler
	PanelLoans     *handler.WebPanelLoansHandler
	PanelReports   *handler.WebPanelReportsHandler
	PanelUsers     *handler.WebPanelUsersHandler
	Import         *handler.WebPanelImportHandler
}

// buildWeb creates the renderer, the web repositories and all web handlers.
func buildWeb(db *gorm.DB, sec securityHandlers) (webHandlers, error) {
	// 1. cfg, err := loadWebConfig(). On error return webHandlers{}, err.
	cfg, err := loadWebConfig()
	if err != nil {
		return webHandlers{}, err
	}
	// 2. render, err := handler.NewRenderer(web.FS, cfg.Location). On error return webHandlers{}, err.
	render, err := handler.NewRenderer(web.FS, cfg.Location)
	if err != nil {
		return webHandlers{}, err
	}
	// 3. catalog := repository.NewCatalogRepository(db); dict := repository.NewDictionaryRepository(db); panel := repository.NewPanelRepository(db); views := repository.NewLoanViewRepository(db); reports := repository.NewReportRepository(db).
	catalog := repository.NewCatalogRepository(db)
	dict := repository.NewDictionaryRepository(db)
	panel := repository.NewPanelRepository(db)
	views := repository.NewLoanViewRepository(db)
	reports := repository.NewReportRepository(db)
	// 4. Return webHandlers{
	//        Render: render, Cookies: cfg.Cookies, Sec: sec,
	//        Auth:           handler.NewWebAuthHandler(sec.AuthService, render, cfg.Cookies),
	//        Catalog:        handler.NewWebCatalogHandler(catalog, dict, render),
	//        Loans:          handler.NewWebLoansHandler(views, sec.Policy, render),
	//        PanelDashboard: handler.NewWebPanelDashboardHandler(reports, render),
	//        PanelBooks:     handler.NewWebPanelBooksHandler(panel, catalog, dict, render),
	//        PanelDicts:     handler.NewWebPanelDictionariesHandler(dict, render),
	//        PanelLoans:     handler.NewWebPanelLoansHandler(panel, views, sec.LoanService, sec.Users, sec.Policy, render),
	//        PanelReports:   handler.NewWebPanelReportsHandler(reports, sec.Policy, render),
	//        PanelUsers:     handler.NewWebPanelUsersHandler(sec.Users, sec.Refresh, render),
	//    }, nil.
	importRepo := repository.NewImportRepository(db)
	exportRepo := repository.NewExportRepository(db)
	importer := service.NewImportService(importRepo)
	return webHandlers{
		Render:         render,
		Cookies:        cfg.Cookies,
		Sec:            sec,
		Auth:           handler.NewWebAuthHandler(sec.AuthService, render, cfg.Cookies),
		Catalog:        handler.NewWebCatalogHandler(catalog, dict, render),
		Loans:          handler.NewWebLoansHandler(views, sec.Policy, render),
		PanelDashboard: handler.NewWebPanelDashboardHandler(reports, render),
		PanelBooks:     handler.NewWebPanelBooksHandler(panel, catalog, dict, render),
		PanelDicts:     handler.NewWebPanelDictionariesHandler(dict, render),
		PanelLoans:     handler.NewWebPanelLoansHandler(panel, views, sec.LoanService, sec.Users, sec.Policy, render),
		PanelReports:   handler.NewWebPanelReportsHandler(reports, sec.Policy, render),
		PanelUsers:     handler.NewWebPanelUsersHandler(sec.Users, sec.Refresh, sec.AuthService, render),
		Import:         handler.NewWebPanelImportHandler(importer, exportRepo, render),
	}, nil
}
