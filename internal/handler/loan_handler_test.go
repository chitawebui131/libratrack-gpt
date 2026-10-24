package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type fakeVerifier struct{ id uint }

func (f fakeVerifier) Verify(token string) (uint, string, error) {
	return f.id, model.RoleLibrarian, nil
}

type fakeLoanService struct {
	issueErr   error
	returnErr  error
	lastUserID uint
	lastStatus string
}

func (f *fakeLoanService) Issue(copyID, userID uint) (*model.Loan, error) {
	if f.issueErr != nil {
		return nil, f.issueErr
	}
	return &model.Loan{ID: 1, CopyID: copyID, UserID: userID}, nil
}

func (f *fakeLoanService) Return(loanID uint) (*model.Loan, error) {
	if f.returnErr != nil {
		return nil, f.returnErr
	}
	return &model.Loan{ID: loanID, FineAmount: 4}, nil
}

func (f *fakeLoanService) List(status string, page, limit int) ([]model.Loan, error) {
	f.lastStatus = status
	return []model.Loan{}, nil
}

func (f *fakeLoanService) ListByUser(userID uint, page, limit int) ([]model.Loan, error) {
	f.lastUserID = userID
	return []model.Loan{}, nil
}

// newLoanTestRouter registers POST /loans, POST /loans/:id/return, GET /loans, GET /me/loans
// behind middleware.AuthMiddleware(fakeVerifier{id: 42}). Every request must send "Authorization: Bearer x".
func newLoanTestRouter(svc LoanService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AuthMiddleware(fakeVerifier{id: 42}))
	h := NewLoanHandler(svc)
	r.POST("/loans", h.Create)
	r.POST("/loans/:id/return", h.Return)
	r.GET("/loans", h.List)
	r.GET("/me/loans", h.ListByUser)
	return r
}

func TestCreateLoanOK(t *testing.T) {
	svc := &fakeLoanService{}
	router := newLoanTestRouter(svc)

	body := `{"copy_id":1,"user_id":2}`
	req := httptest.NewRequest("POST", "/loans", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestCreateLoanCopyNotAvailable(t *testing.T) {
	svc := &fakeLoanService{issueErr: repository.ErrCopyNotAvailable}
	router := newLoanTestRouter(svc)

	body := `{"copy_id":1,"user_id":2}`
	req := httptest.NewRequest("POST", "/loans", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if code, ok := resp["code"].(string); !ok || code != "copy_not_available" {
		t.Errorf("expected code 'copy_not_available', got %v", resp["code"])
	}
}

func TestCreateLoanLimitReached(t *testing.T) {
	svc := &fakeLoanService{issueErr: repository.ErrLoanLimitReached}
	router := newLoanTestRouter(svc)

	body := `{"copy_id":1,"user_id":2}`
	req := httptest.NewRequest("POST", "/loans", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if code, ok := resp["code"].(string); !ok || code != "loan_limit_reached" {
		t.Errorf("expected code 'loan_limit_reached', got %v", resp["code"])
	}
}

func TestReturnLoanOK(t *testing.T) {
	svc := &fakeLoanService{}
	router := newLoanTestRouter(svc)

	req := httptest.NewRequest("POST", "/loans/5/return", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestReturnLoanAlreadyReturned(t *testing.T) {
	svc := &fakeLoanService{returnErr: repository.ErrAlreadyReturned}
	router := newLoanTestRouter(svc)

	req := httptest.NewRequest("POST", "/loans/5/return", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
}

func TestReturnLoanNotFound(t *testing.T) {
	svc := &fakeLoanService{returnErr: repository.ErrNotFound}
	router := newLoanTestRouter(svc)

	req := httptest.NewRequest("POST", "/loans/5/return", nil)
	req.Header.Set("Authorization", "Bearer x")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
