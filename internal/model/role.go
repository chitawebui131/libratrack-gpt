package model

const (
	RoleAdmin     = "admin"
	RoleLibrarian = "librarian"
	RoleReader    = "reader"
)

// IsValidRole reports whether role is one of the known roles.
func IsValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleLibrarian, RoleReader:
		return true
	}
	return false
}
