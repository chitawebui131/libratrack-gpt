package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
)

// registerWebRoutes registers /static, every web page route and the NoRoute handler.
func registerWebRoutes(r *gin.Engine, w webHandlers) error {
	// 1. static, err := staticHandler(). If err != nil return err. r.GET("/static/*filepath", static); r.HEAD("/static/*filepath", static).
	static, err := staticHandler()
	if err != nil {
		return err
	}
	r.GET("/static/*filepath", static)
	r.HEAD("/static/*filepath", static)

	// 2. chain := []gin.HandlerFunc{
	//        middleware.SecurityHeadersMiddleware(),
	//        middleware.CSRFMiddleware(w.Cookies, w.Render.Forbidden),
	//        middleware.WebAuthMiddleware(w.Sec.TokenManager, w.Sec.AuthService, w.Sec.Users, w.Cookies),
	//    }.
	chain := []gin.HandlerFunc{
		middleware.SecurityHeadersMiddleware(),
		middleware.CSRFMiddleware(w.Cookies, w.Render.Forbidden),
		middleware.WebAuthMiddleware(w.Sec.TokenManager, w.Sec.AuthService, w.Sec.Users, w.Cookies),
	}

	// 3. web := r.Group("", chain...).
	limit := middleware.BodyLimitMiddleware(10 << 20)
	web := r.Group("", append([]gin.HandlerFunc{limit}, chain...)...)

	// 4. Public: web.GET("/", w.Auth.Index); web.GET("/login", w.Auth.LoginPage); web.POST("/login", w.Auth.Login); web.GET("/register", w.Auth.RegisterPage); web.POST("/register", w.Auth.Register); web.POST("/logout", w.Auth.Logout).
	web.GET("/", w.Auth.Index)
	web.GET("/login", w.Auth.LoginPage)
	web.POST("/login", w.Auth.Login)
	web.GET("/register", w.Auth.RegisterPage)
	web.POST("/register", w.Auth.Register)
	web.POST("/logout", w.Auth.Logout)
	web.GET("/catalog", w.Catalog.Catalog)
	web.GET("/books/:id", w.Catalog.Book)

	// 5. Public: the catalog and the book page. Logged-in only: my loans.
	authed := web.Group("", middleware.RequireLogin())
	authed.GET("/my/loans", w.Loans.MyLoans)

	// 6. staff := web.Group("", middleware.RequireWebRoles(w.Render.Forbidden, model.RoleLibrarian, model.RoleAdmin)):
	//      GET /panel -> w.PanelDashboard.Dashboard
	//      GET /panel/new-book -> w.PanelBooks.NewBookPage; POST /panel/books -> CreateBook; GET /panel/books/:id/edit -> EditBookPage; POST /panel/books/:id -> UpdateBook; POST /panel/books/:id/delete -> DeleteBook
	//      GET /panel/authors -> w.PanelDicts.AuthorsPage; POST /panel/authors -> CreateAuthor; GET /panel/authors/:id/edit -> EditAuthorPage; POST /panel/authors/:id -> UpdateAuthor
	//      GET /panel/categories -> CategoriesPage; POST /panel/categories -> CreateCategory; GET /panel/categories/:id/edit -> EditCategoryPage; POST /panel/categories/:id -> UpdateCategory
	//      POST /panel/books/:id/copies -> w.PanelLoans.AddCopy; POST /panel/copies/:id/status -> SetCopyStatus
	//      GET /panel/loans -> LoansPage; POST /panel/loans -> IssueLoan; POST /panel/loans/:id/return -> ReturnLoan
	//      GET /panel/reports -> w.PanelReports.Reports
	staff := web.Group("", middleware.RequireWebRoles(w.Render.Forbidden, model.RoleLibrarian, model.RoleAdmin))
	staff.GET("/panel", w.PanelDashboard.Dashboard)
	staff.GET("/panel/books", w.PanelBooks.BooksPage)
	staff.GET("/panel/new-book", w.PanelBooks.NewBookPage)
	staff.POST("/panel/books", w.PanelBooks.CreateBook)
	staff.GET("/panel/books/:id/edit", w.PanelBooks.EditBookPage)
	staff.POST("/panel/books/:id", w.PanelBooks.UpdateBook)
	staff.POST("/panel/books/:id/delete", w.PanelBooks.DeleteBook)
	staff.GET("/panel/authors", w.PanelDicts.AuthorsPage)
	staff.POST("/panel/authors", w.PanelDicts.CreateAuthor)
	staff.GET("/panel/authors/:id/edit", w.PanelDicts.EditAuthorPage)
	staff.POST("/panel/authors/:id", w.PanelDicts.UpdateAuthor)
	staff.POST("/panel/authors/:id/delete", w.PanelDicts.DeleteAuthor)
	staff.GET("/panel/categories", w.PanelDicts.CategoriesPage)
	staff.POST("/panel/categories", w.PanelDicts.CreateCategory)
	staff.GET("/panel/categories/:id/edit", w.PanelDicts.EditCategoryPage)
	staff.POST("/panel/categories/:id", w.PanelDicts.UpdateCategory)
	staff.POST("/panel/categories/:id/delete", w.PanelDicts.DeleteCategory)
	staff.POST("/panel/books/:id/copies", w.PanelLoans.AddCopy)
	staff.POST("/panel/copies/:id/status", w.PanelLoans.SetCopyStatus)
	staff.POST("/panel/copies/:id", w.PanelLoans.UpdateCopy)
	staff.POST("/panel/copies/:id/delete", w.PanelLoans.DeleteCopy)
	staff.GET("/panel/loans", w.PanelLoans.LoansPage)
	staff.POST("/panel/loans", w.PanelLoans.IssueLoan)
	staff.POST("/panel/loans/:id/return", w.PanelLoans.ReturnLoan)
	staff.GET("/panel/reports", w.PanelReports.Reports)
	staff.GET("/panel/import", w.Import.ImportPage)
	staff.POST("/panel/import", w.Import.ImportUpload)
	staff.GET("/panel/export", w.Import.ExportFile)

	// 7. admin := web.Group("", middleware.RequireWebRoles(w.Render.Forbidden, model.RoleAdmin)): GET /panel/users -> w.PanelUsers.UsersPage; POST /panel/users/:id/role -> ChangeRole.
	admin := web.Group("", middleware.RequireWebRoles(w.Render.Forbidden, model.RoleAdmin))
	admin.GET("/panel/users", w.PanelUsers.UsersPage)
	admin.POST("/panel/users/:id/role", w.PanelUsers.ChangeRole)
  admin.GET("/panel/new-user", w.PanelUsers.NewUserPage)
  admin.POST("/panel/users", w.PanelUsers.CreateUser)
  admin.GET("/panel/users/:id/edit", w.PanelUsers.EditUserPage)
  admin.POST("/panel/users/:id", w.PanelUsers.UpdateUser)
  admin.POST("/panel/users/:id/disable", w.PanelUsers.SetDisabled)
  admin.POST("/panel/users/:id/delete", w.PanelUsers.DeleteUser)
	// 8. r.NoRoute(
	//        func(c *gin.Context) {
	//            if strings.HasPrefix(c.Request.URL.Path, "/api/") {
	//                c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "not_found", "message": "route not found", "details": ""}})
	//            }
	//        },
	//        chain[0], chain[1], chain[2],
	//        w.Render.NotFound,
	//    ).
	r.NoRoute(
		func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "not_found", "message": "route not found", "details": ""}})
			}
		},
		limit,
		chain[0], chain[1], chain[2],
		w.Render.NotFound,
	)

	// 9. Return nil.
	return nil
}
