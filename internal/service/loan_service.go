package service

import (
	"time"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type LoanService struct {
	loans  repository.LoanRepository
	policy LoanPolicy
	now    func() time.Time // tests may replace it
}

func NewLoanService(loans repository.LoanRepository, policy LoanPolicy) *LoanService {
	// Return &LoanService{loans: loans, policy: policy, now: time.Now}.
	return &LoanService{loans: loans, policy: policy, now: time.Now}
}

func (s *LoanService) Issue(copyID, userID uint) (*model.Loan, error) {
	// 1. borrowedAt := s.now().UTC() and dueAt := s.policy.DueDate(borrowedAt).
	// 2. Return s.loans.Issue(copyID, userID, borrowedAt, dueAt, s.policy.MaxActiveLoans).
	borrowedAt := s.now().UTC()
	dueAt := s.policy.DueDate(borrowedAt)
	return s.loans.Issue(copyID, userID, borrowedAt, dueAt, s.policy.MaxActiveLoans)
}

func (s *LoanService) Return(loanID uint) (*model.Loan, error) {
	// Return s.loans.Return(loanID, s.now().UTC(), s.policy.CalculateFine).
	return s.loans.Return(loanID, s.now().UTC(), s.policy.CalculateFine)
}

func (s *LoanService) List(status string, page, limit int) ([]model.Loan, error) {
	// Return s.loans.List(status, page, limit).
	return s.loans.List(status, page, limit)
}

func (s *LoanService) ListByUser(userID uint, page, limit int) ([]model.Loan, error) {
	// Return s.loans.ListByUser(userID, page, limit).
	return s.loans.ListByUser(userID, page, limit)
}
