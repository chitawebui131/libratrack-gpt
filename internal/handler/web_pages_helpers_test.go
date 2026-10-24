package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/repository"
	"libratrack/web"
)

// pagesVerifier accepts the token "reader-token" (user 1) and "librarian-token" (user 2).
type pagesVerifier struct{}

func (pagesVerifier) Verify(token string) (uint, string, error) {
	switch token {
	case "reader-token":
		return 1, "reader", nil
	case "librarian-token":
		return 2, "librarian", nil
	default:
		return 0, "", errors.New("bad token")
	}
}

type pagesUsers struct{}

func (pagesUsers) FindByID(id uint) (*model.User, error) {
	switch id {
	case 1:
		return &model.User{ID: 1, Email: "r@x.com", FullName: "Reader", Role: model.RoleReader}, nil
	case 2:
		return &model.User{ID: 2, Email: "l@x.com", FullName: "Librarian", Role: model.RoleLibrarian}, nil
	default:
		return nil, repository.ErrNotFound
	}
}

type pagesRefresher struct{}

func (pagesRefresher) Refresh(token string) (*model.AuthResponse, error) {
	return nil, repository.ErrNotFound
}

type fakeCatalog struct {
	repository.CatalogRepository
	book *model.Book
}

func (f fakeCatalog) Search(filter repository.BookFilter) ([]model.Book, int64, error) {
	return nil, 0, nil
}
func (f fakeCatalog) AvailabilityForBooks(ids []uint) (map[uint]model.CatalogAvailability, error) {
	return map[uint]model.CatalogAvailability{}, nil
}
func (f fakeCatalog) GetBook(id uint) (*model.Book, error) {
	if f.book == nil || f.book.ID != id {
		return nil, repository.ErrNotFound
	}
	return f.book, nil
}

type fakeDict struct {
	repository.DictionaryRepository
}

func (fakeDict) ListCategories() ([]model.Category, error) { return nil, nil }
func (fakeDict) ListAuthors() ([]model.Author, error)      { return nil, nil }

type fakeReports struct{ repository.ReportRepository }

func (fakeReports) Stats(now time.Time) (*model.DashboardStats, error) {
	return &model.DashboardStats{}, nil
}

type fakePanel struct{ repository.PanelRepository }

// newPagesRouter assembles the router: SecurityHeaders, CSRF, WebAuth, then the routes
// GET /catalog (RequireLogin), GET /books/:id (RequireLogin), GET /panel (staff), POST /panel/books (staff),
// and GET /api/ping behind middleware.AuthMiddleware(pagesVerifier{}).
func newPagesRouter(t *testing.T, book *model.Book) *gin.Engine {
	render, err := NewRenderer(web.FS, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	cookies := middleware.CookieConfig{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	web := r.Group("", middleware.SecurityHeadersMiddleware(), middleware.CSRFMiddleware(cookies, render.Forbidden), middleware.WebAuthMiddleware(pagesVerifier{}, pagesRefresher{}, pagesUsers{}, cookies))
	catalog := NewWebCatalogHandler(fakeCatalog{book: book}, fakeDict{}, render)
	// 1. Register GET /catalog and GET /books/:id on the public web group (without RequireLogin)
	web.GET("/catalog", catalog.Catalog)
	web.GET("/books/:id", catalog.Book)
	// 2. Register staff routes and /api/ping with appropriate middleware
	staff := web.Group("", middleware.RequireWebRoles(render.Forbidden, model.RoleLibrarian, model.RoleAdmin))
	staff.GET("/panel", NewWebPanelDashboardHandler(fakeReports{}, render).Dashboard)
	staff.POST("/panel/books", NewWebPanelBooksHandler(fakePanel{}, fakeCatalog{}, fakeDict{}, render).CreateBook)
	r.GET("/api/ping", middleware.AuthMiddleware(pagesVerifier{}), func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// do performs a request. cookies is a list of "name=value" strings joined into the Cookie header. form may be nil.
func do(r *gin.Engine, method, path string, cookies []string, form url.Values) *httptest.ResponseRecorder {
	var reqBody string
	var contentType string
	if form != nil {
		reqBody = form.Encode()
		contentType = "application/x-www-form-urlencoded"
	}
	req := httptest.NewRequest(method, path, strings.NewReader(reqBody))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if len(cookies) > 0 {
		req.Header.Set("Cookie", strings.Join(cookies, "; "))
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// csrfCookie performs GET /catalog as the reader to get a CSRF cookie value.
func csrfCookie(t *testing.T, r *gin.Engine) string {
	w := do(r, "GET", "/catalog", []string{"access_token=reader-token"}, nil)
	for _, c := range w.Result().Cookies() {
		if c.Name == "csrf_token" {
			return c.Value
		}
	}
	t.Fatal("CSRF cookie not found")
	return ""
}
