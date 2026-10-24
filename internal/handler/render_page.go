package handler

import (
	"bytes"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
)

// Render writes the page. page is the template file name without extension, for example "catalog".
func (r *Renderer) Render(c *gin.Context, status int, page, title string, data any) {
	// 1. t, ok := r.pages[page]. If !ok: log.Printf("unknown page %q", page); c.String(http.StatusInternalServerError, "internal server error"); return.
	t, ok := r.pages[page]
	if !ok {
		log.Printf("unknown page %q", page)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}
	// 2. code, arg := ReadFlash(c).
	code, arg := ReadFlash(c)
	// 3. user, _ := middleware.WebUser(c).
	user, _ := middleware.WebUser(c)
	// 4. pd := PageData{Title: title, User: user, CSRFToken: middleware.CSRFToken(c), Flash: code, FlashArg: arg, Data: data}.
	pd := PageData{Title: title, User: user, CSRFToken: middleware.CSRFToken(c), Flash: code, FlashArg: arg, Data: data}
	// 5. var buf bytes.Buffer. If err := t.ExecuteTemplate(&buf, "layout", pd); err != nil: log.Printf("template error: %v", err); c.String(http.StatusInternalServerError, "internal server error"); return.
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template error: %v", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}
	// 6. c.Data(status, "text/html; charset=utf-8", buf.Bytes()).
	c.Data(status, "text/html; charset=utf-8", buf.Bytes())
}

func (r *Renderer) errorPage(c *gin.Context, status int, title string) {
	r.Render(c, status, "error", title, ErrorPage{Code: status})
	c.Abort()
}

func (r *Renderer) Forbidden(c *gin.Context)   { r.errorPage(c, http.StatusForbidden, "403") }
func (r *Renderer) NotFound(c *gin.Context)    { r.errorPage(c, http.StatusNotFound, "404") }
func (r *Renderer) ServerError(c *gin.Context) { r.errorPage(c, http.StatusInternalServerError, "500") }
