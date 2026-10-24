package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"libratrack/web"
)

// TestTemplatesParse proves that every template in web.FS parses.
// It is the safety net for every later template step.
func TestTemplatesParse(t *testing.T) {
	// 1. _, err := NewRenderer(web.FS, time.UTC). If err != nil, t.Fatalf("templates do not parse: %v", err).
	_, err := NewRenderer(web.FS, time.UTC)
	if err != nil {
		t.Fatalf("templates do not parse: %v", err)
	}
}

func TestTemplateFuncs(t *testing.T) {
	// 1. f := templateFuncs(time.UTC).
	f := templateFuncs(time.UTC)
	// 2. date: f["date"].(func(time.Time) string) for 2025-03-05 10:20 UTC returns "2025-03-05"; datetime returns "2025-03-05 10:20".
	date := f["date"].(func(time.Time) string)
	datetime := f["datetime"].(func(time.Time) string)
	tm := time.Date(2025, 3, 5, 10, 20, 0, 0, time.UTC)
	if date(tm) != "2025-03-05" {
		t.Errorf("date function failed: expected '2025-03-05', got '%s'", date(tm))
	}
	if datetime(tm) != "2025-03-05 10:20" {
		t.Errorf("datetime function failed: expected '2025-03-05 10:20', got '%s'", datetime(tm))
	}
	// 3. statusClass: f["statusClass"].(func(any) string)("under_repair") returns "repair"; with "bogus" returns "".
	statusClass := f["statusClass"].(func(any) string)
	if statusClass("under_repair") != "repair" {
		t.Errorf("statusClass function failed for 'under_repair': expected 'repair', got '%s'", statusClass("under_repair"))
	}
	if statusClass("bogus") != "" {
		t.Errorf("statusClass function failed for 'bogus': expected '', got '%s'", statusClass("bogus"))
	}
}

func TestErrorPage404(t *testing.T) {
	// 1. r, err := NewRenderer(web.FS, time.UTC). Fatal on error.
	r, err := NewRenderer(web.FS, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	// 2. gin.SetMode(gin.TestMode); w := httptest.NewRecorder(); c, _ := gin.CreateTestContext(w); c.Request = httptest.NewRequest(http.MethodGet, "/x", nil).
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	// 3. r.NotFound(c). The status must be 404 and the body must contain the ASCII marker "page-title".
	r.NotFound(c)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "page-title") {
		t.Errorf("response body does not contain 'page-title': %s", w.Body.String())
	}
}

func TestFlashInvalidCodeIgnored(t *testing.T) {
	// 1. Create a test context like above. Call SetFlash(c, "Bad Code!", 0). The Set-Cookie header of the recorder must be empty.
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	SetFlash(c, "Bad Code!", 0)
	cookieHeader := w.Header().Get("Set-Cookie")
	if cookieHeader != "" {
		t.Errorf("expected empty Set-Cookie header, got %q", cookieHeader)
	}
}
