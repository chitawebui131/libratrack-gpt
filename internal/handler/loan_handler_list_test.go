package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeLoanService is a mock implementation of LoanService for testing

// newLoanTestRouter registers POST /loans, POST /loans/:id/return, GET /loans, GET /me/loans
// behind middleware.AuthMiddleware(fakeVerifier{id: 42}). Every request must send "Authorization: Bearer x".

func TestListLoansInvalidStatus(t *testing.T) {
	// GET /loans?status=bogus with header "Authorization: Bearer x" -> 422.
	svc := &fakeLoanService{}
	router := newLoanTestRouter(svc)

	req := httptest.NewRequest("GET", "/loans?status=bogus", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}

func TestListLoansValidStatus(t *testing.T) {
	// GET /loans?status=active -> 200 and the fake saw lastStatus == "active".
	svc := &fakeLoanService{}
	router := newLoanTestRouter(svc)

	req := httptest.NewRequest("GET", "/loans?status=active", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if svc.lastStatus != "active" {
		t.Errorf("expected lastStatus to be 'active', got %s", svc.lastStatus)
	}
}

func TestListMyLoansUsesTokenUser(t *testing.T) {
	// GET /me/loans?user_id=999 -> 200 and the fake saw lastUserID == 42 (the id from the token, not from the query).
	svc := &fakeLoanService{}
	router := newLoanTestRouter(svc)

	req := httptest.NewRequest("GET", "/me/loans?user_id=999", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if svc.lastUserID != 42 {
		t.Errorf("expected lastUserID to be 42, got %d", svc.lastUserID)
	}
}
