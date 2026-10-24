package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
)

// newGuardRouter returns a router where setWebUser is called first when role != "" and then handler middleware runs.
func newGuardRouter(role string, guard gin.HandlerFunc) *gin.Engine {
	// 1. gin.SetMode(gin.TestMode); r := gin.New().
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 2. If role != "": r.Use(func(c *gin.Context) { setWebUser(c, &model.User{ID: 1, Role: role}); c.Next() }).
	if role != "" {
		r.Use(func(c *gin.Context) {
			setWebUser(c, &model.User{ID: 1, Role: role})
			c.Next()
		})
	}
	// 3. r.GET("/catalog", guard, func(c *gin.Context) { c.Status(http.StatusOK) }).
	r.GET("/catalog", guard, func(c *gin.Context) { c.Status(http.StatusOK) })
	// 4. Return r.
	return r
}

func forbidden(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) }

func TestRequireLoginRedirects(t *testing.T) {
	// GET /catalog?x=1 without a user -> 303 and Location "/login?next=%2Fcatalog%3Fx%3D1".
	r := newGuardRouter("", RequireLogin())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/catalog?x=1", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	expectedLocation := "/login?next=%2Fcatalog%3Fx%3D1"
	if w.Header().Get("Location") != expectedLocation {
		t.Errorf("Expected location %s, got %s", expectedLocation, w.Header().Get("Location"))
	}
}

func TestRequireWebRoles(t *testing.T) {
	// guard := RequireWebRoles(forbidden, model.RoleLibrarian, model.RoleAdmin).
	// reader -> 403; librarian -> 200; admin -> 200.
	guard := RequireWebRoles(forbidden, model.RoleLibrarian, model.RoleAdmin)
	r := newGuardRouter(model.RoleReader, guard)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/catalog", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	r = newGuardRouter(model.RoleLibrarian, guard)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/catalog", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	r = newGuardRouter(model.RoleAdmin, guard)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/catalog", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	// Router with SecurityHeadersMiddleware and GET /x. The response has non-empty Content-Security-Policy, X-Frame-Options "DENY" and X-Content-Type-Options "nosniff".
	r := gin.New()
	r.Use(SecurityHeadersMiddleware())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/x", nil)
	r.ServeHTTP(w, req)
	if w.Header().Get("Content-Security-Policy") == "" {
		t.Errorf("Expected Content-Security-Policy header")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options header to be DENY")
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options header to be nosniff")
	}
}
