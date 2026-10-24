package handler

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const flashCookie = "flash"

var flashCodeRe = regexp.MustCompile(`^[a-z_]+$`)

// SetFlash stores a flash message for the next page. Invalid codes are ignored.
func SetFlash(c *gin.Context, code string, arg int64) {
	if !flashCodeRe.MatchString(code) {
		return
	}
	value := code
	if arg > 0 {
		value = code + ":" + strconv.FormatInt(arg, 10)
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: flashCookie, Value: value, Path: "/", MaxAge: 60, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

// ReadFlash returns the flash message and deletes the cookie.
func ReadFlash(c *gin.Context) (code string, arg int64) {
	ck, err := c.Request.Cookie(flashCookie)
	if err != nil {
		return "", 0
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: flashCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	parts := strings.SplitN(ck.Value, ":", 2)
	if !flashCodeRe.MatchString(parts[0]) {
		return "", 0
	}
	if len(parts) == 2 {
		n, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || n < 0 {
			return parts[0], 0
		}
		return parts[0], n
	}
	return parts[0], 0
}

// RedirectWithFlash sets a flash message and redirects with 303.
func RedirectWithFlash(c *gin.Context, location, code string, arg int64) {
	SetFlash(c, code, arg)
	c.Redirect(http.StatusSeeOther, location)
}
