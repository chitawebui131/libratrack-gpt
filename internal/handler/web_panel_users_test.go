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

func (usersTestRefresher) RevokeAllForUser(userID uint) error {
	return nil
}

type fakeDisableCall struct {
	id       uint
	disabled bool
}

// fakeUsersRepo records the calls. It must be shared by pointer between the router and the test.
type fakeUsersRepo struct {
	repository.UserRepository
	deleteCalls      []uint
	setDisabledCalls []fakeDisableCall
}

func (f *fakeUsersRepo) Delete(id uint) error {
	f.deleteCalls = append(f.deleteCalls, id)
	return nil
}

func (f *fakeUsersRepo) SetDisabled(id uint, disabled bool) error {
	f.setDisabledCalls = append(f.setDisabledCalls, fakeDisableCall{id: id, disabled: disabled})
	return nil
}

type fakeAdmin struct{}

func (fakeAdmin) CreateUser(email, password, fullName, role string) (*model.User, error) {
	return &model.User{Email: email, FullName: fullName, Role: role}, nil
}

func (fakeAdmin) ResetPassword(userID uint, newPassword string) error {
	return nil
}

func (fakeAdmin) UpdateProfile(userID uint, email, fullName string) error {
	return nil
}

// usersTestRouter returns the router and the ONE fake repository used by all handlers.
func usersTestRouter(t *testing.T) (*gin.Engine, *fakeUsersRepo) {
	t.Helper()
	render, err := NewRenderer(web.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	cookies := middleware.CookieConfig{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := &fakeUsersRepo{}
	h := NewWebPanelUsersHandler(repo, usersTestRefresher{}, fakeAdmin{}, render)

	g := r.Group("", middleware.SecurityHeadersMiddleware(), middleware.CSRFMiddleware(cookies, render.Forbidden), middleware.WebAuthMiddleware(usersTestVerifier{}, usersTestRefresher{}, usersTestUsers{}, cookies))
	admin := g.Group("", middleware.RequireWebRoles(render.Forbidden, model.RoleAdmin))
	admin.GET("/panel/new-user", h.NewUserPage)
	admin.POST("/panel/users/:id/delete", h.DeleteUser)
	admin.POST("/panel/users/:id/disable", h.SetDisabled)
	return r, repo
}

func usersTestCSRF(t *testing.T, r *gin.Engine) string {
	t.Helper()
	w := do(r, "GET", "/panel/new-user", []string{"access_token=admin-token"}, nil)
	for _, c := range w.Result().Cookies() {
		if c.Name == "csrf_token" {
			return c.Value
		}
	}
	t.Fatal("CSRF cookie not found")
	return ""
}

func usersTestFlash(w interface{ Result() *http.Response }) string {
	for _, c := range w.Result().Cookies() {
		if c.Name == "flash" {
			return c.Value
		}
	}
	return ""
}

func TestNewUserPageIsAdminOnly(t *testing.T) {
	r, _ := usersTestRouter(t)

	w := do(r, "GET", "/panel/new-user", []string{}, nil)
	if w.Code != http.StatusSeeOther {
		t.Errorf("guest: expected status %d, got %d", http.StatusSeeOther, w.Code)
	}

	w = do(r, "GET", "/panel/new-user", []string{"access_token=librarian-token"}, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("librarian: expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	w = do(r, "GET", "/panel/new-user", []string{"access_token=admin-token"}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("admin: expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `name="password"`) {
		t.Errorf("expected the form to contain name=\"password\"")
	}
	if !strings.Contains(body, `name="role"`) {
		t.Errorf("expected the form to contain name=\"role\"")
	}
}

func TestAdminCannotDeleteOrDisableSelf(t *testing.T) {
	r, repo := usersTestRouter(t)
	csrf := usersTestCSRF(t, r)
	cookies := []string{"access_token=admin-token", "csrf_token=" + csrf}

	// Delete self (admin is user 3).
	w := do(r, "POST", "/panel/users/3/delete", cookies, url.Values{"csrf_token": {csrf}})
	if w.Code != http.StatusSeeOther {
		t.Errorf("delete self: expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/panel/users" {
		t.Errorf("delete self: expected redirect to /panel/users, got %s", loc)
	}
	if f := usersTestFlash(w); f != "user_self_action" {
		t.Errorf("delete self: expected flash user_self_action, got %q", f)
	}

	// Disable self.
	w = do(r, "POST", "/panel/users/3/disable", cookies, url.Values{"csrf_token": {csrf}, "disabled": {"1"}})
	if w.Code != http.StatusSeeOther {
		t.Errorf("disable self: expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/panel/users" {
		t.Errorf("disable self: expected redirect to /panel/users, got %s", loc)
	}
	if f := usersTestFlash(w); f != "user_self_action" {
		t.Errorf("disable self: expected flash user_self_action, got %q", f)
	}

	if len(repo.deleteCalls) != 0 {
		t.Errorf("expected no delete calls, got %v", repo.deleteCalls)
	}
	if len(repo.setDisabledCalls) != 0 {
		t.Errorf("expected no set-disabled calls, got %v", repo.setDisabledCalls)
	}
}

func TestAdminDeletesAnotherUser(t *testing.T) {
	r, repo := usersTestRouter(t)
	csrf := usersTestCSRF(t, r)
	cookies := []string{"access_token=admin-token", "csrf_token=" + csrf}

	w := do(r, "POST", "/panel/users/7/delete", cookies, url.Values{"csrf_token": {csrf}})
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/panel/users" {
		t.Errorf("expected redirect to /panel/users, got %s", loc)
	}
	if f := usersTestFlash(w); f != "user_deleted" {
		t.Errorf("expected flash user_deleted, got %q", f)
	}
	if len(repo.deleteCalls) != 1 || repo.deleteCalls[0] != 7 {
		t.Fatalf("expected one delete call with id 7, got %v", repo.deleteCalls)
	}
}