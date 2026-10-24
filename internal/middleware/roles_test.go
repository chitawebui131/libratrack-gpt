package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newRolesRouter returns a router where GET /admin requires the role "admin".
// If withAuth is true, AuthMiddleware(fakeVerifier{id: 1, role: role}) runs before RequireRoles.
func newRolesRouter(withAuth bool, role string) *gin.Engine {
	// 1. gin.SetMode(gin.TestMode); r := gin.New().
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 2. If withAuth: r.Use(AuthMiddleware(fakeVerifier{id: 1, role: role})).
	if withAuth {
		r.Use(AuthMiddleware(fakeVerifier{id: 1, role: role}))
	}
	// 3. r.GET("/admin", RequireRoles("admin"), func(c *gin.Context) { c.Status(http.StatusOK) }).
	r.GET("/admin", RequireRoles("admin"), func(c *gin.Context) { c.Status(http.StatusOK) })
	// 4. Return r.
	return r
}

func TestRequireRolesAllowed(t *testing.T) {
	// withAuth true, role "admin", header "Bearer x" -> 200.
	router := newRolesRouter(true, "admin")
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRequireRolesForbidden(t *testing.T) {
	// withAuth true, role "reader", header "Bearer x" -> 403 and JSON error.code "forbidden".
	router := newRolesRouter(true, "reader")
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	expectedBody := `{"error":{"code":"forbidden","message":"insufficient permissions","details":""}}`
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestRequireRolesWithoutUser(t *testing.T) {
	// withAuth false -> 401.
	router := newRolesRouter(false, "")
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
