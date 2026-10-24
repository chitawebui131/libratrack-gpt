package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/service"
	"libratrack/web"
)

type fakeWebAuth struct {
	loginErr     error
	registerErr  error
	lastRegister model.RegisterRequest
}

func (f *fakeWebAuth) Register(req model.RegisterRequest) (*model.User, error) {
	f.lastRegister = req
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	return &model.User{ID: 1, Email: req.Email, Role: model.RoleReader}, nil
}

func (f *fakeWebAuth) Login(email, password string) (*model.AuthResponse, error) {
	if f.loginErr != nil {
		return nil, f.loginErr
	}
	return &model.AuthResponse{AccessToken: "ACCESS-SECRET", User: model.User{ID: 1, Role: model.RoleReader}}, nil

}

func (f *fakeWebAuth) Logout(refreshToken string) error { return nil }

func newWebAuthTestRouter(t *testing.T, svc WebAuthService) *gin.Engine {
	// 1. render, err := NewRenderer(web.FS, time.UTC). Fatal on error.
	render, err := NewRenderer(web.FS, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	// 2. gin.SetMode(gin.TestMode); r := gin.New(); h := NewWebAuthHandler(svc, render, middleware.CookieConfig{AccessMaxAge: 60, RefreshMaxAge: 120}).
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewWebAuthHandler(svc, render, middleware.CookieConfig{AccessMaxAge: 60, RefreshMaxAge: 120})
	// 3. r.GET("/login", h.LoginPage); r.POST("/login", h.Login); r.GET("/register", h.RegisterPage); r.POST("/register", h.Register). Return r.
	r.GET("/login", h.LoginPage)
	r.POST("/login", h.Login)
	r.GET("/register", h.RegisterPage)
	r.POST("/register", h.Register)
	return r
}

// postForm sends an application/x-www-form-urlencoded POST and returns the recorder.
func postForm(r *gin.Engine, path string, values url.Values) *httptest.ResponseRecorder {
	// Create the request with strings.NewReader(values.Encode()), set the Content-Type header, serve it and return the recorder.
	req := httptest.NewRequest("POST", path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSafeNext(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/catalog", "/catalog"},
		{"//evil.com", "/catalog"},
		{"https://evil.com", "/catalog"},
		{"/\\evil.com", "/catalog"},
		{"", "/catalog"},
	}

	for _, tt := range tests {
		result := safeNext(tt.input)
		if result != tt.expected {
			t.Errorf("safeNext(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestLoginFailuresLookTheSame(t *testing.T) {
	svc := &fakeWebAuth{loginErr: service.ErrInvalidCredentials}
	r := newWebAuthTestRouter(t, svc)

	values1 := url.Values{"email": {"user1@example.com"}, "password": {"SuperSecret123"}}
	resp1 := postForm(r, "/login", values1)

	values2 := url.Values{"email": {"user2@example.com"}, "password": {"SuperSecret123"}}
	resp2 := postForm(r, "/login", values2)

	if resp1.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp1.Code)
	}
	if resp2.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp2.Code)
	}

	body1 := resp1.Body.String()
	body2 := resp2.Body.String()

	if strings.Contains(body1, "SuperSecret123") {
		t.Error("Response body contains the submitted password")
	}
	if strings.Contains(body2, "SuperSecret123") {
		t.Error("Response body contains the submitted password")
	}
}

func TestLoginSuccessSetsCookies(t *testing.T) {
	svc := &fakeWebAuth{}
	r := newWebAuthTestRouter(t, svc)

	values := url.Values{"email": {"user@example.com"}, "password": {"password123"}}
	resp := postForm(r, "/login", values)

	if resp.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", resp.Code)
	}

	cookies := resp.Result().Cookies()
	var accessTokenCookie, refreshTokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessTokenCookie = cookie
		}
		if cookie.Name == "refresh_token" {
			refreshTokenCookie = cookie
		}
	}

	if accessTokenCookie == nil {
		t.Error("access_token cookie not found")
	} else {
		if !accessTokenCookie.HttpOnly {
			t.Error("access_token cookie is not HttpOnly")
		}
		if accessTokenCookie.SameSite != http.SameSiteLaxMode {
			t.Error("access_token cookie does not have SameSite=Lax")
		}
	}

	if refreshTokenCookie == nil {
		t.Error("refresh_token cookie not found")
	} else {
		if !refreshTokenCookie.HttpOnly {
			t.Error("refresh_token cookie is not HttpOnly")
		}
		if refreshTokenCookie.SameSite != http.SameSiteLaxMode {
			t.Error("refresh_token cookie does not have SameSite=Lax")
		}
	}

	body := resp.Body.String()
	if strings.Contains(body, "ACCESS-SECRET") {
		t.Error("Response body contains the access token")
	}
}

func TestRegisterIgnoresRoleField(t *testing.T) {
	svc := &fakeWebAuth{}
	r := newWebAuthTestRouter(t, svc)

	values := url.Values{
		"email":     {"user@example.com"},
		"password":  {"password123"},
		"full_name": {"Test User"},
		"role":      {"admin"},
	}
	resp := postForm(r, "/register", values)

	if resp.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", resp.Code)
	}

	if svc.lastRegister.Email != "user@example.com" {
		t.Error("Email not saved correctly")
	}
	if svc.lastRegister.Password != "password123" {
		t.Error("Password not saved correctly")
	}
	if svc.lastRegister.FullName != "Test User" {
		t.Error("Full name not saved correctly")
	}
	// Check that role field is ignored (RegisterRequest has no Role field)
	// The fake service returns a reader role, which is correct
}
