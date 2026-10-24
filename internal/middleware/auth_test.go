package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeVerifier struct {
	id   uint
	role string
	err  error
}

func (f fakeVerifier) Verify(token string) (uint, string, error) {
	return f.id, f.role, f.err
}

// newAuthRouter returns a router with AuthMiddleware and one route GET /ping that answers 200
// and writes the id from CurrentUser into the response body.
func newAuthRouter(v TokenVerifier) *gin.Engine {
	// 1. gin.SetMode(gin.TestMode); r := gin.New().
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 2. r.Use(AuthMiddleware(v)).
	r.Use(AuthMiddleware(v))
	// 3. r.GET("/ping", func(c *gin.Context) { id, _, _ := CurrentUser(c); c.JSON(http.StatusOK, gin.H{"id": id}) }).
	r.GET("/ping", func(c *gin.Context) { id, _, _ := CurrentUser(c); c.JSON(http.StatusOK, gin.H{"id": id}) })
	// 4. Return r.
	return r
}

func TestAuthNoHeader(t *testing.T) {
	// GET /ping without Authorization -> 401, header WWW-Authenticate is "Bearer", JSON body has error.code "unauthorized".
	router := newAuthRouter(fakeVerifier{})
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	if w.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Errorf("expected WWW-Authenticate header to be 'Bearer', got '%s'", w.Header().Get("WWW-Authenticate"))
	}

	// Check the JSON body
	expectedBody := `{"error":{"code":"unauthorized","message":"authentication required","details":""}}`
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestAuthWrongScheme(t *testing.T) {
	// Authorization "Basic abc" -> 401.
	router := newAuthRouter(fakeVerifier{})
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Basic abc")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthVerifierRejects(t *testing.T) {
	// Verifier returns errors.New("bad"); header "Bearer x" -> 401.
	router := newAuthRouter(fakeVerifier{err: errors.New("bad")})
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthValidToken(t *testing.T) {
	// Verifier returns id 5, role "reader"; header "Bearer x" -> 200 and the body contains id 5.
	router := newAuthRouter(fakeVerifier{id: 5, role: "reader"})
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	expectedBody := `{"id":5}`
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
	}
}
