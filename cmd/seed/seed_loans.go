package main

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"libratrack/internal/model"
	"libratrack/internal/repository"
	"libratrack/internal/service"
)

// createLoans creates one active loan, one overdue loan and one returned loan with a fine.
func createLoans(db *gorm.DB, loans repository.LoanRepository, policy service.LoanPolicy, readers map[string]uint) error {
	// 1. var copies []model.Copy. err := db.Where("status = ?", model.CopyStatusAvailable).Order("id").Limit(3).Find(&copies).Error. Return the error if any.
	var copies []model.Copy
	err := db.Where("status = ?", model.CopyStatusAvailable).Order("id").Limit(3).Find(&copies).Error
	if err != nil {
		return err
	}
	// 2. If len(copies) < 3: fmt.Println("warning: fewer than 3 available copies, creating as many loans as possible").
	if len(copies) < 3 {
		fmt.Println("warning: fewer than 3 available copies, creating as many loans as possible")
	}
	// 3. now := time.Now().UTC(); max := policy.MaxActiveLoans.
	now := time.Now().UTC()
	max := policy.MaxActiveLoans
	// 4. If len(copies) > 0: loan A: loans.Issue(copies[0].ID, readers["reader1@libratrack.local"], now, policy.DueDate(now), max). Return the error if any. Print "loan A (active)".
	if len(copies) > 0 {
		_, err := loans.Issue(copies[0].ID, readers["reader1@libratrack.local"], now, policy.DueDate(now), max)
		if err != nil {
			return err
		}
		fmt.Println("loan A (active)")
	}
	// 5. If len(copies) > 1: loan B: borrowed := now.AddDate(0, 0, -30); due := now.AddDate(0, 0, -16); loans.Issue(copies[1].ID, readers["reader2@libratrack.local"], borrowed, due, max). Return the error if any. Print "loan B (overdue)".
	if len(copies) > 1 {
		borrowed := now.AddDate(0, 0, -30)
		due := now.AddDate(0, 0, -16)
		_, err := loans.Issue(copies[1].ID, readers["reader2@libratrack.local"], borrowed, due, max)
		if err != nil {
			return err
		}
		fmt.Println("loan B (overdue)")
	}
	// 6. If len(copies) > 2: loan C: borrowed := now.AddDate(0, 0, -40); due := now.AddDate(0, 0, -26); l, err := loans.Issue(copies[2].ID, readers["reader2@libratrack.local"], borrowed, due, max). Then loans.Return(l.ID, now.AddDate(0, 0, -23), policy.CalculateFine). Return the errors if any. Print "loan C (returned with a fine)".
	if len(copies) > 2 {
		borrowed := now.AddDate(0, 0, -40)
		due := now.AddDate(0, 0, -26)
		l, err := loans.Issue(copies[2].ID, readers["reader2@libratrack.local"], borrowed, due, max)
		if err != nil {
			return err
		}
		_, err = loans.Return(l.ID, now.AddDate(0, 0, -23), policy.CalculateFine)
		if err != nil {
			return err
		}
		fmt.Println("loan C (returned with a fine)")
	}
	// 7. Return nil.
	return nil
}
