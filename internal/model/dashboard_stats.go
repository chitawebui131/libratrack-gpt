package model

// DashboardStats represents statistics for the librarian dashboard.
type DashboardStats struct {
	TotalBooks      int `json:"total_books"`
	TotalCopies     int `json:"total_copies"`
	AvailableCopies int `json:"available_copies"`
	ActiveLoans     int `json:"active_loans"`
	OverdueLoans    int `json:"overdue_loans"`
}
