package main

import (
	"errors"

	"libratrack/internal/model"
	"libratrack/internal/service"
)

// seedUsers creates the admin and the optional demo users. It never changes existing users.
func seedUsers(auth *service.AuthService, cfg securityConfig) error {
	// 1. If cfg.AdminEmail != "" and cfg.AdminPassword != "": call auth.EnsureUser(cfg.AdminEmail, cfg.AdminPassword, "Administrator", model.RoleAdmin). Return the error if any.
	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		if err := auth.EnsureUser(cfg.AdminEmail, cfg.AdminPassword, "Administrator", model.RoleAdmin); err != nil {
			return err
		}
	}
	// 2. If cfg.SeedDemoUsers:
	//      if cfg.DemoUsersPassword == "": return errors.New("DEMO_USERS_PASSWORD is required when SEED_DEMO_USERS=true").
	if cfg.SeedDemoUsers {
		if cfg.DemoUsersPassword == "" {
			return errors.New("DEMO_USERS_PASSWORD is required when SEED_DEMO_USERS=true")
		}
		if err := auth.EnsureUser("librarian@libratrack.local", cfg.DemoUsersPassword, "Demo Librarian", model.RoleLibrarian); err != nil {
			return err
		}
		if err := auth.EnsureUser("reader@libratrack.local", cfg.DemoUsersPassword, "Demo Reader", model.RoleReader); err != nil {
			return err
		}
	}
	// 3. Return nil.
	return nil
}
