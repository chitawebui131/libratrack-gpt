package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCSRFRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CSRFMiddleware(CookieConfig{}, func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) }))
	r.GET("/form", func(c *gin.Context) { c.String(http.StatusOK, CSRFToken(c)) })
	r.POST("/form", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// getCookie performs GET /form and returns the value of the csrf_token cookie.
func getCookie(t *testing.T, r *gin.Engine) string {
	req := httptest.NewRequest("GET", "/form", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == CSRFCookieName {
			return cookie.Value
		}
	}
	t.Fatalf("CSRF cookie not found")
	return ""
}

func TestCSRFGetSetsCookie(t *testing.T) {
	r := newCSRFRouter()
	cookie := getCookie(t, r)
	if cookie == "" {
		t.Errorf("Expected non-empty CSRF cookie")
	}
}

func TestCSRFPostWithoutToken(t *testing.T) {
	r := newCSRFRouter()
	cookie := getCookie(t, r)

	req := httptest.NewRequest("POST", "/form", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: cookie})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", w.Code)
	}
}

func TestCSRFPostWrongToken(t *testing.T) {
	r := newCSRFRouter()
	cookie := getCookie(t, r)

	form := url.Values{}
	form.Add("csrf_token", "wrong_token")
	req := httptest.NewRequest("POST", "/form", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: cookie})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", w.Code)
	}
}

func TestCSRFPostMatchingToken(t *testing.T) {
	r := newCSRFRouter()
	cookie := getCookie(t, r)

	form := url.Values{}
	form.Add("csrf_token", cookie)
	req := httptest.NewRequest("POST", "/form", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: cookie})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}
