package main

import (
	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
	"libratrack/internal/model"
)

// registerSecurityRoutes registers the new routes on api (the /api/v1 group)
// and returns the three protected groups so that main.go can put the old routes into them.
func registerSecurityRoutes(api *gin.RouterGroup, h securityHandlers) (authed, staff, admin *gin.RouterGroup) {
	// 1. Public: api.POST("/auth/register", h.Auth.Register) and api.POST("/auth/login", h.Auth.Login).
	api.POST("/auth/register", h.Auth.Register)
	api.POST("/auth/login", h.Auth.Login)
	// 2. authed = api.Group("", middleware.AuthMiddleware(h.TokenManager)).
	authed = api.Group("", middleware.AuthMiddleware(h.TokenManager))
	// 3. staff = authed.Group("", middleware.RequireRoles(model.RoleLibrarian, model.RoleAdmin)).
	staff = authed.Group("", middleware.RequireRoles(model.RoleLibrarian, model.RoleAdmin))
	// 4. admin = authed.Group("", middleware.RequireRoles(model.RoleAdmin)).
	admin = authed.Group("", middleware.RequireRoles(model.RoleAdmin))
	// 5. authed.GET("/me", h.Auth.Me) and authed.GET("/me/loans", h.Loan.ListMine).
	authed.GET("/me", h.Auth.Me)
	authed.GET("/me/loans", h.Loan.ListMine)
	// 6. staff.POST("/loans", h.Loan.Create), staff.POST("/loans/:id/return", h.Loan.Return), staff.GET("/loans", h.Loan.List).
	staff.POST("/loans", h.Loan.Create)
	staff.POST("/loans/:id/return", h.Loan.Return)
	staff.GET("/loans", h.Loan.List)
	// 7. admin.GET("/users", h.User.List) and admin.PUT("/users/:id/role", h.User.UpdateRole).
	admin.GET("/users", h.User.List)
	admin.PUT("/users/:id/role", h.User.UpdateRole)
	// 8. Return authed, staff, admin.
	return authed, staff, admin
}
