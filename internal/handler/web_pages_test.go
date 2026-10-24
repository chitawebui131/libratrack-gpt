package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"libratrack/internal/model"
)

func TestAnonymousCatalogIsPublic(t *testing.T) {
	r := newPagesRouter(t, nil)
	w := do(r, "GET", "/catalog", []string{}, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAnonymousBookPageHidesInventory(t *testing.T) {
	book := &model.Book{ID: 5, Title: "Test Book", Copies: []model.Copy{{ID: 1, InventoryNo: "INV-SECRET-1", Status: model.CopyStatusAvailable}}}
	r := newPagesRouter(t, book)
	w := do(r, "GET", "/books/5", []string{}, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if strings.Contains(body, "INV-SECRET-1") {
		t.Errorf("Expected body to not contain 'INV-SECRET-1', got %s", body)
	}
}

func TestAnonymousPanelRedirects(t *testing.T) {
	r := newPagesRouter(t, nil)
	w := do(r, "GET", "/panel", []string{}, nil)
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
	}
	location := w.Result().Header.Get("Location")
	if !strings.HasPrefix(location, "/login") {
		t.Errorf("Expected location to start with '/login', got %s", location)
	}
}

func TestReaderCannotOpenPanel(t *testing.T) {
	r := newPagesRouter(t, nil)
	w := do(r, "GET", "/panel", []string{"access_token=reader-token"}, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "page-title") {
		t.Errorf("Expected body to contain 'page-title', got %s", body)
	}
}

func TestReaderPostWithValidCSRFStillForbidden(t *testing.T) {
	r := newPagesRouter(t, nil)
	csrf := csrfCookie(t, r)
	form := url.Values{"csrf_token": {csrf}}
	w := do(r, "POST", "/panel/books", []string{"access_token=reader-token", "csrf_token=" + csrf}, form)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestLibrarianPanelOK(t *testing.T) {
	r := newPagesRouter(t, nil)
	w := do(r, "GET", "/panel", []string{"access_token=librarian-token"}, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestPostWithoutCSRFForbidden(t *testing.T) {
	r := newPagesRouter(t, nil)
	w := do(r, "POST", "/panel/books", []string{"access_token=librarian-token"}, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestUnknownBookIs404(t *testing.T) {
	r := newPagesRouter(t, nil)
	w := do(r, "GET", "/books/999", []string{"access_token=reader-token"}, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestReaderDoesNotSeeInventoryNumbers(t *testing.T) {
	book := &model.Book{ID: 5, Title: "Test Book", Copies: []model.Copy{{ID: 1, InventoryNo: "INV-SECRET-1", Status: model.CopyStatusAvailable}}}
	r := newPagesRouter(t, book)

	// Test as reader
	w := do(r, "GET", "/books/5", []string{"access_token=reader-token"}, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if strings.Contains(body, "INV-SECRET-1") {
		t.Errorf("Expected body to not contain 'INV-SECRET-1', got %s", body)
	}

	// Test as librarian
	w = do(r, "GET", "/books/5", []string{"access_token=librarian-token"}, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body = w.Body.String()
	if !strings.Contains(body, "INV-SECRET-1") {
		t.Errorf("Expected body to contain 'INV-SECRET-1', got %s", body)
	}
}

func TestAPIDoesNotAcceptCookies(t *testing.T) {
	r := newPagesRouter(t, nil)

	// Test with cookie
	w := do(r, "GET", "/api/ping", []string{"access_token=reader-token"}, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test with Authorization header
	req := httptest.NewRequest("GET", "/api/ping", nil)
	req.Header.Set("Authorization", "Bearer reader-token")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
