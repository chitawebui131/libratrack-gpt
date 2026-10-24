package repository

import (
	"time"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

// ReportRepository reads data for the librarian dashboard.
type ReportRepository interface {
	Stats(now time.Time) (*model.DashboardStats, error)

	OverdueLoans(now time.Time) ([]model.LoanView, error)
	TopBooks(limit int) ([]model.TopBook, error)
}

type gormReportRepository struct {
	db *gorm.DB
}

// NewReportRepository creates a new instance of ReportRepository.
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &gormReportRepository{db: db}
}
