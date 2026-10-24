package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

const ctxCSRF = "csrf_token_value"

// CSRFToken returns the CSRF token of the current request (for hidden form fields).
func CSRFToken(c *gin.Context) string {
	v, _ := c.Get(ctxCSRF)
	s, _ := v.(string)
	return s
}

func newCSRFToken() (string, error) {
	// 1. b := make([]byte, 32); rand.Read(b). Return the error if any.
	// 2. Return base64.RawURLEncoding.EncodeToString(b), nil.
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// CSRFMiddleware protects every non-safe request with a double-submit cookie.
func CSRFMiddleware(cfg CookieConfig, onFail gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. cookieToken, err := c.Cookie(CSRFCookieName). If err != nil or cookieToken == "":
		//      generate a new token with newCSRFToken (on error call onFail(c); c.Abort(); return);
		//      set the cookie with http.SetCookie(c.Writer, &http.Cookie{Name: CSRFCookieName, Value: token, Path: "/", HttpOnly: true, Secure: cfg.Secure, SameSite: http.SameSiteLaxMode});
		//      then set cookieToken = "" for the comparison below and remember the new token as the one in the context.
		cookieToken, err := c.Cookie(CSRFCookieName)
		if err != nil || cookieToken == "" {
			token, err := newCSRFToken()
			if err != nil {
				onFail(c)
				c.Abort()
				return
			}
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     CSRFCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   cfg.Secure,
				SameSite: http.SameSiteLaxMode,
			})
			cookieToken = ""
		}
		// 2. c.Set(ctxCSRF, the token of this request).
		c.Set(ctxCSRF, cookieToken)
		// 3. If the method is GET, HEAD or OPTIONS: c.Next() and return.
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		// 4. submitted := c.PostForm("csrf_token"). If submitted == "", submitted = c.GetHeader("X-CSRF-Token").
		submitted := c.PostForm("csrf_token")
		if submitted == "" {
			submitted = c.GetHeader("X-CSRF-Token")
		}
		// 5. If cookieToken == "" or submitted == "" or subtle.ConstantTimeCompare([]byte(submitted), []byte(cookieToken)) != 1: onFail(c); c.Abort(); return.
		if cookieToken == "" || submitted == "" || subtle.ConstantTimeCompare([]byte(submitted), []byte(cookieToken)) != 1 {
			onFail(c)
			c.Abort()
			return
		}
		// 6. c.Next().
		c.Next()
	}
}
