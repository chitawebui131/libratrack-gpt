package main

import (
	"errors"
	"fmt"

	"libratrack/internal/model"
	"libratrack/internal/repository"
	"libratrack/internal/service"
)

type demoUser struct {
	Email string
	Name  string
	Role  string
}

var demoUsers = []demoUser{
	{"admin@libratrack.local", "Demo Admin", model.RoleAdmin},
	{"librarian@libratrack.local", "Demo Librarian", model.RoleLibrarian},
	{"reader1@libratrack.local", "Demo Reader One", model.RoleReader},
	{"reader2@libratrack.local", "Demo Reader Two", model.RoleReader},
}

// createUsers creates the demo users and skips existing emails.
// It returns the ids by email and whether reader1 already existed.
func createUsers(users repository.UserRepository, password string) (ids map[string]uint, reader1Existed bool, err error) {
	// 1. ids = map[string]uint{}. hash, err := service.HashPassword(password). Return the error if any.
	ids = map[string]uint{}
	hash, err := service.HashPassword(password)
	if err != nil {
		return nil, false, err
	}
	// 2. For every u in demoUsers:
	//      existing, err := users.FindByEmail(u.Email).
	//      If err == nil: ids[u.Email] = existing.ID; if u.Email == "reader1@libratrack.local" { reader1Existed = true }; fmt.Println("skipped (exists):", u.Email); continue.
	//      If !errors.Is(err, repository.ErrNotFound): return nil, false, err.
	//      nu := &model.User{Email: u.Email, FullName: u.Name, PasswordHash: hash, Role: u.Role}. If err := users.Create(nu); err != nil return nil, false, err.
	//      ids[u.Email] = nu.ID; fmt.Println("created:", u.Email).
	for _, u := range demoUsers {
		existing, err := users.FindByEmail(u.Email)
		if err == nil {
			ids[u.Email] = existing.ID
			if u.Email == "reader1@libratrack.local" {
				reader1Existed = true
			}
			fmt.Println("skipped (exists):", u.Email)
			continue
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, false, err
		}
		nu := &model.User{Email: u.Email, FullName: u.Name, PasswordHash: hash, Role: u.Role}
		if err := users.Create(nu); err != nil {
			return nil, false, err
		}
		ids[u.Email] = nu.ID
		fmt.Println("created:", u.Email)
	}
	// 3. Return ids, reader1Existed, nil.
	return ids, reader1Existed, nil
}
