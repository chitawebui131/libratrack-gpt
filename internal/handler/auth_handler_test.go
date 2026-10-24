package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/service"
)

type fakeAuthService struct {
	registerErr error
	loginErr    error
}

func (f *fakeAuthService) Register(req model.RegisterRequest) (*model.User, error) {
	// If f.registerErr != nil return nil, f.registerErr. Otherwise return &model.User{ID: 1, Email: req.Email, Role: model.RoleReader}, nil.
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	return &model.User{ID: 1, Email: req.Email, Role: model.RoleReader}, nil
}

func (f *fakeAuthService) Login(email, password string) (*model.AuthResponse, error) {
	// If f.loginErr != nil return nil, f.loginErr. Otherwise return &model.AuthResponse{AccessToken: "tok", TokenType: "Bearer", ExpiresIn: 60}, nil.
	if f.loginErr != nil {
		return nil, f.loginErr
	}
	return &model.AuthResponse{AccessToken: "tok", TokenType: "Bearer", ExpiresIn: 60}, nil
}

func (f *fakeAuthService) GetByID(id uint) (*model.User, error) {
	// Return &model.User{ID: id}, nil.
	return &model.User{ID: id}, nil
}

// newAuthTestRouter returns a router with POST /auth/register and POST /auth/login.
func newAuthTestRouter(svc AuthService) *gin.Engine {
	// gin.SetMode(gin.TestMode); r := gin.New(); h := NewAuthHandler(svc); register the two routes; return r.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler(svc)
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	return r
}

func TestRegisterOK(t *testing.T) {
	// Valid body {"email":"a@b.com","password":"secret123","full_name":"A B"} -> 201. The body must NOT contain "password" and must NOT contain "password_hash".
	router := newAuthTestRouter(&fakeAuthService{})
	body := `{"email":"a@b.com","password":"secret123","full_name":"A B"}`
	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
	respBody := w.Body.String()
	if strings.Contains(respBody, "password") {
		t.Error("response body must not contain password")
	}
	if strings.Contains(respBody, "password_hash") {
		t.Error("response body must not contain password_hash")
	}
}

func TestRegisterShortPassword(t *testing.T) {
	// password "short" -> 422.
	router := newAuthTestRouter(&fakeAuthService{})
	body := `{"email":"a@b.com","password":"short","full_name":"A B"}`
	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	// fake registerErr = service.ErrEmailTaken -> 409.
	router := newAuthTestRouter(&fakeAuthService{registerErr: service.ErrEmailTaken})
	body := `{"email":"a@b.com","password":"secret123","full_name":"A B"}`
	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
}

func TestLoginOK(t *testing.T) {
	// Valid body {"email":"a@b.com","password":"secret123"} -> 200 and the body contains "access_token".
	router := newAuthTestRouter(&fakeAuthService{})
	body := `{"email":"a@b.com","password":"secret123"}`
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	respBody := w.Body.String()
	if !strings.Contains(respBody, "access_token") {
		t.Error("response body must contain access_token")
	}
}

func TestLoginWrongCredentials(t *testing.T) {
	// fake loginErr = service.ErrInvalidCredentials -> 401.
	router := newAuthTestRouter(&fakeAuthService{loginErr: service.ErrInvalidCredentials})
	body := `{"email":"a@b.com","password":"secret123"}`
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestLoginInvalidBody(t *testing.T) {
	// body "{}" -> 422.
	router := newAuthTestRouter(&fakeAuthService{})
	body := `{}`
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}
