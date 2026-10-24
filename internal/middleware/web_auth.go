package middleware

import (
	"github.com/gin-gonic/gin"
)

func WebAuthMiddleware(v TokenVerifier, refresher WebRefresher, users WebUserFinder, cfg CookieConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. access, err := c.Cookie(AccessCookieName). If err == nil and access != "":
		//      id, _, err := v.Verify(access). If err == nil:
		//        u, err := users.FindByID(id). If err == nil: setWebUser(c, u); c.Next(); return.
		access, err := c.Cookie(AccessCookieName)
		if err == nil && access != "" {
			id, _, err := v.Verify(access)
			if err == nil {
				u, err := users.FindByID(id)
				if err == nil {
					if u.Disabled {
						ClearAuthCookies(c, cfg)
						c.Next()
						return
					}
					setWebUser(c, u)
					c.Next()
					return
				}
			}
		}
		// 2. refresh, err := c.Cookie(RefreshCookieName). If err == nil and refresh != "":
		//      resp, err := refresher.Refresh(refresh).
		//      If err == nil: SetAuthCookies(c, cfg, resp); user := resp.User; setWebUser(c, &user); c.Next(); return.
		//      Otherwise: ClearAuthCookies(c, cfg).
		refresh, err := c.Cookie(RefreshCookieName)
		if err == nil && refresh != "" {
			resp, err := refresher.Refresh(refresh)
			if err == nil {
				SetAuthCookies(c, cfg, resp)
				user := resp.User
				setWebUser(c, &user)
				c.Next()
				return
			} else {
				ClearAuthCookies(c, cfg)
			}
		}
		// 3. c.Next().
		c.Next()
	}
}
