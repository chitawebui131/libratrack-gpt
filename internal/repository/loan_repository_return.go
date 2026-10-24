package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"libratrack/internal/model"
)

func (r *gormLoanRepository) Return(loanID uint, returnedAt time.Time, calcFine func(dueAt, returnedAt time.Time) int) (*model.Loan, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. var loan model.Loan. Load it: tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&loan, loanID).
		//    If errors.Is(err, gorm.ErrRecordNotFound), return ErrNotFound. For any other error return it.
		var loan model.Loan
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&loan, loanID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		// 2. If loan.ReturnedAt != nil, return ErrAlreadyReturned.
		if loan.ReturnedAt != nil {
			return ErrAlreadyReturned
		}

		// 3. fine := calcFine(loan.DueAt, returnedAt).
		fine := calcFine(loan.DueAt, returnedAt)

		// 4. Update the loan: tx.Model(&model.Loan{}).Where("id = ?", loanID).Updates(map[string]interface{}{"returned_at": returnedAt, "fine_amount": fine}). Return the error if any.
		err = tx.Model(&model.Loan{}).Where("id = ?", loanID).Updates(map[string]interface{}{"returned_at": returnedAt, "fine_amount": fine}).Error
		if err != nil {
			return err
		}

		// 5. Update the copy: tx.Model(&model.Copy{}).Where("id = ?", loan.CopyID).Updates(map[string]interface{}{"status": "available", "updated_at": time.Now()}). Return the error if any.
		err = tx.Model(&model.Copy{}).Where("id = ?", loan.CopyID).Updates(map[string]interface{}{"status": "available", "updated_at": time.Now()}).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, translateLoanError(err)
	}
	// 6. Reload: var out model.Loan; err = r.db.Preload("Copy.Book").First(&out, loanID).Error.
	//    If err != nil return nil, translateLoanError(err). Otherwise return &out, nil.
	var out model.Loan
	err = r.db.Preload("Copy.Book").First(&out, loanID).Error
	if err != nil {
		return nil, translateLoanError(err)
	}
	return &out, nil
}
