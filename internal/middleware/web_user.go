package middleware

import (
	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
)

const ctxWebUser = "web_user"

// WebRefresher rotates a refresh token. *service.AuthService implements it.
type WebRefresher interface {
	Refresh(refreshToken string) (*model.AuthResponse, error)
}

// WebUserFinder loads a user. repository.UserRepository implements it.
type WebUserFinder interface {
	FindByID(id uint) (*model.User, error)
}

func setWebUser(c *gin.Context, u *model.User) {
	c.Set(ctxWebUser, u)
}

// WebUser returns the logged-in user of a web request.
func WebUser(c *gin.Context) (*model.User, bool) {
	// 1. v, exists := c.Get(ctxWebUser). If !exists, return nil, false.
	// 2. u, ok := v.(*model.User). If !ok or u == nil, return nil, false.
	// 3. Return u, true.
	v, exists := c.Get(ctxWebUser)
	if !exists {
		return nil, false
	}
	u, ok := v.(*model.User)
	if !ok || u == nil {
		return nil, false
	}
	return u, true
}
