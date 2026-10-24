package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
)

type fakeUsers struct {
	user *model.User
}

func (f fakeUsers) FindByID(id uint) (*model.User, error) {
	if f.user == nil {
		return nil, errors.New("not found")
	}
	return f.user, nil
}

type fakeRefresher struct {
	resp *model.AuthResponse
	err  error
}

func (f fakeRefresher) Refresh(token string) (*model.AuthResponse, error) {
	return f.resp, f.err
}

// newWebAuthRouter has WebAuthMiddleware and GET /who, which answers "anon" or the role of the user from WebUser.
func newWebAuthRouter(v TokenVerifier, r WebRefresher, u WebUserFinder) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.Use(WebAuthMiddleware(v, r, u, CookieConfig{AccessMaxAge: 60, RefreshMaxAge: 120}))
	e.GET("/who", func(c *gin.Context) {
		if u, _ := WebUser(c); u != nil {
			c.String(http.StatusOK, u.Role)
		} else {
			c.String(http.StatusOK, "anon")
		}
	})
	return e
}

func TestWebAuthAnonymous(t *testing.T) {
	// No cookies -> body "anon".
	router := newWebAuthRouter(fakeVerifier{}, fakeRefresher{}, fakeUsers{})
	req := httptest.NewRequest("GET", "/who", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "anon" {
		t.Errorf("expected body 'anon', got '%s'", w.Body.String())
	}
}

func TestWebAuthRoleFromDatabase(t *testing.T) {
	// Verifier returns (1, "reader", nil) but the fake users return a user with Role "librarian". With the access cookie set, the body must be "librarian".
	router := newWebAuthRouter(fakeVerifier{id: 1, role: "reader"}, fakeRefresher{}, fakeUsers{user: &model.User{ID: 1, Role: "librarian"}})
	req := httptest.NewRequest("GET", "/who", nil)
	req.Header.Set("Cookie", "access_token=abc")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "librarian" {
		t.Errorf("expected body 'librarian', got '%s'", w.Body.String())
	}
}

func TestWebAuthRefreshRotates(t *testing.T) {
	// Verifier returns an error. Refresher returns an AuthResponse with AccessToken "a2", RefreshToken "r2" and User{ID: 1, Role: "reader"}. Cookies: access_token=old, refresh_token=oldr.
	// The body is "reader". The response has Set-Cookie headers for access_token and refresh_token, each containing "HttpOnly" and "SameSite=Lax".
	router := newWebAuthRouter(fakeVerifier{err: errors.New("invalid token")}, fakeRefresher{resp: &model.AuthResponse{
		AccessToken: "a2",
		User:        model.User{ID: 1, Role: "reader"},
	}}, fakeUsers{})
	req := httptest.NewRequest("GET", "/who", nil)
	req.Header.Set("Cookie", "access_token=old; refresh_token=oldr")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "reader" {
		t.Errorf("expected body 'reader', got '%s'", w.Body.String())
	}

	// Check Set-Cookie headers
	setCookies := w.Header().Values("Set-Cookie")
	if len(setCookies) < 2 {
		t.Error("expected at least 2 Set-Cookie headers")
	}

	hasAccessCookie := false
	hasRefreshCookie := false

	for _, cookie := range setCookies {
		if strings.Contains(cookie, "access_token=") && strings.Contains(cookie, "HttpOnly") && strings.Contains(cookie, "SameSite=Lax") {
			hasAccessCookie = true
		}
		if strings.Contains(cookie, "refresh_token=") && strings.Contains(cookie, "HttpOnly") && strings.Contains(cookie, "SameSite=Lax") {
			hasRefreshCookie = true
		}
	}

	if !hasAccessCookie {
		t.Error("expected access token cookie with HttpOnly and SameSite=Lax")
	}
	if !hasRefreshCookie {
		t.Error("expected refresh token cookie with HttpOnly and SameSite=Lax")
	}
}

func TestWebAuthInvalidClearsCookies(t *testing.T) {
	// Verifier and refresher both return errors. Body "anon". The Set-Cookie headers clear both cookies (they contain "Max-Age=0").
	router := newWebAuthRouter(fakeVerifier{err: errors.New("invalid token")}, fakeRefresher{err: errors.New("refresh failed")}, fakeUsers{})
	req := httptest.NewRequest("GET", "/who", nil)
	req.Header.Set("Cookie", "access_token=old; refresh_token=oldr")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "anon" {
		t.Errorf("expected body 'anon', got '%s'", w.Body.String())
	}

	// Check Set-Cookie headers
	setCookies := w.Header().Values("Set-Cookie")
	if len(setCookies) < 2 {
		t.Error("expected at least 2 Set-Cookie headers")
	}

	hasAccessCookie := false
	hasRefreshCookie := false

	for _, cookie := range setCookies {
		if strings.Contains(cookie, "access_token=") && strings.Contains(cookie, "Max-Age=0") {
			hasAccessCookie = true
		}
		if strings.Contains(cookie, "refresh_token=") && strings.Contains(cookie, "Max-Age=0") {
			hasRefreshCookie = true
		}
	}

	if !hasAccessCookie {
		t.Error("expected access token cookie with Max-Age=0")
	}
	if !hasRefreshCookie {
		t.Error("expected refresh token cookie with Max-Age=0")
	}
}
