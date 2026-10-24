package handler

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/repository"
	"libratrack/web"
)

// usersTestVerifier accepts "admin-token" (user 3) and "librarian-token" (user 2).
type usersTestVerifier struct{}

func (usersTestVerifier) Verify(token string) (uint, string, error) {
	switch token {
	case "admin-token":
		return 3, "admin", nil
	case "librarian-token":
		return 2, "librarian", nil
	default:
		return 0, "", repository.ErrNotFound
	}
}

type usersTestUsers struct{}

func (usersTestUsers) FindByID(id uint) (*model.User, error) {
	switch id {
	case 2:
		return &model.User{ID: 2, Email: "l@x.com", FullName: "Librarian", Role: model.RoleLibrarian}, nil
	case 3:
		return &model.User{ID: 3, Email: "a@x.com", FullName: "Boss", Role: model.RoleAdmin}, nil
	default:
		return nil, repository.ErrNotFound
	}
}

type usersTestRefresher struct{}

func (usersTestRefresher) Refresh(token string) (*model.AuthResponse, error) {
	return nil, repository.ErrNotFound
}

type fakeUsersRepo struct {
	repository.UserRepository
	deleteCalls      []uint
	setDisabledCalls []struct {
		id       uint
		disabled bool
	}
}
func (usersTestRefresher) RevokeAllForUser(userID uint) error {
	return nil
}
func (f *fakeUsersRepo) Delete(id uint) error {
	f.deleteCalls = append(f.deleteCalls, id)
	return nil
}

func (f *fakeUsersRepo) SetDisabled(id uint, disabled bool) error {
	f.setDisabledCalls = append(f.setDisabledCalls, struct {
		id       uint
		disabled bool
	}{id: id, disabled: disabled})
	return nil
}

type fakeAdmin struct{}

func (fakeAdmin) Get() error {
	// This method is not used by the tests
	return nil
}

func (fakeAdmin) CreateUser(email, password, fullName, role string) (*model.User, error) {
	return &model.User{
		Email:    email,
		FullName: fullName,
		Role:     role,
	}, nil
}

func (fakeAdmin) ResetPassword(userID uint, newPassword string) error {
	return nil
}

func (fakeAdmin) UpdateProfile(userID uint, email, fullName string) error {
	return nil
}

func usersTestRouter(t *testing.T) *gin.Engine {
	render, err := NewRenderer(web.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	cookies := middleware.CookieConfig{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	web := r.Group("", middleware.SecurityHeadersMiddleware(), middleware.CSRFMiddleware(cookies, render.Forbidden), middleware.WebAuthMiddleware(usersTestVerifier{}, usersTestRefresher{}, usersTestUsers{}, cookies))
	admin := web.Group("", middleware.RequireWebRoles(render.Forbidden, model.RoleAdmin))
	admin.GET("/panel/new-user", NewWebPanelUsersHandler(&fakeUsersRepo{}, usersTestRefresher{}, fakeAdmin{}, render).NewUser)
	admin.POST("/panel/users/:id/delete", NewWebPanelUsersHandler(&fakeUsersRepo{}, usersTestRefresher{}, fakeAdmin{}, render).DeleteUser)
	admin.POST("/panel/users/:id/disable", NewWebPanelUsersHandler(&fakeUsersRepo{}, usersTestRefresher{}, fakeAdmin{}, render).SetDisabled)
	return r
}

func TestNewUserPageIsAdminOnly(t *testing.T) {
	r := usersTestRouter(t)

	// Test without cookie
	w := do(r, "GET", "/panel/new-user", []string{}, nil)
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}

	// Test with librarian cookie
	w = do(r, "GET", "/panel/new-user", []string{"access_token=librarian-token"}, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	// Test with admin cookie
	w = do(r, "GET", "/panel/new-user", []string{"access_token=admin-token"}, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `name="password"`) {
		t.Errorf("Expected body to contain 'name=\"password\"', got %s", body)
	}
	if !strings.Contains(body, `name="role"`) {
		t.Errorf("Expected body to contain 'name=\"role\"', got %s", body)
	}
}

func TestAdminCannotDeleteOrDisableSelf(t *testing.T) {
	r := usersTestRouter(t)
	fakeRepo := &fakeUsersRepo{}
	//handler := NewWebPanelUsersHandler(fakeRepo, usersTestRefresher{}, fakeAdmin{}, nil)

	// Test delete self
	csrf := csrfCookie(t, r)
	form := url.Values{"csrf_token": {csrf}}
	w := do(r, "POST", "/panel/users/3/delete", []string{"access_token=admin-token", "csrf_token=" + csrf}, form)
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	location := w.Result().Header.Get("Location")
	if location != "/panel/users" {
		t.Errorf("Expected redirect to '/panel/users', got %s", location)
	}
	flash := ""
	for _, c := range w.Result().Cookies() {
		if c.Name == "flash" {
			flash = c.Value
			break
		}
	}
	if flash != "user_self_action" {
		t.Errorf("Expected flash cookie 'user_self_action', got %s", flash)
	}
	if len(fakeRepo.deleteCalls) != 0 {
		t.Errorf("Expected no delete calls, got %d", len(fakeRepo.deleteCalls))
	}

	// Test disable self
	w = do(r, "POST", "/panel/users/3/disable", []string{"access_token=admin-token", "csrf_token=" + csrf}, form)
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	location = w.Result().Header.Get("Location")
	if location != "/panel/users" {
		t.Errorf("Expected redirect to '/panel/users', got %s", location)
	}
	flash = ""
	for _, c := range w.Result().Cookies() {
		if c.Name == "flash" {
			flash = c.Value
			break
		}
	}
	if flash != "user_self_action" {
		t.Errorf("Expected flash cookie 'user_self_action', got %s", flash)
	}
	if len(fakeRepo.setDisabledCalls) != 0 {
		t.Errorf("Expected no set disabled calls, got %d", len(fakeRepo.setDisabledCalls))
	}
}

func TestAdminDeletesAnotherUser(t *testing.T) {
	r := usersTestRouter(t)
	fakeRepo := &fakeUsersRepo{}

	// Test delete another user
	csrf := csrfCookie(t, r)
	form := url.Values{"csrf_token": {csrf}}
	w := do(r, "POST", "/panel/users/7/delete", []string{"access_token=admin-token", "csrf_token=" + csrf}, form)
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	location := w.Result().Header.Get("Location")
	if location != "/panel/users" {
		t.Errorf("Expected redirect to '/panel/users', got %s", location)
	}
	flash := ""
	for _, c := range w.Result().Cookies() {
		if c.Name == "flash" {
			flash = c.Value
			break
		}
	}
	if flash != "user_deleted" {
		t.Errorf("Expected flash cookie 'user_deleted', got %s", flash)
	}
	if len(fakeRepo.deleteCalls) != 1 {
		t.Errorf("Expected one delete call, got %d", len(fakeRepo.deleteCalls))
	}
	if fakeRepo.deleteCalls[0] != 7 {
		t.Errorf("Expected delete call with id 7, got %d", fakeRepo.deleteCalls[0])
	}
}

func (h *WebPanelUsersHandler) NewUser(c *gin.Context) {
	
	//c.Status(http.StatusOK)
 h.render.Render(c, http.StatusOK, "panel_user_new", "user_new", nil)
}