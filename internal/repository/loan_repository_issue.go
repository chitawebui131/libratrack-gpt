package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"libratrack/internal/model"
)

func (r *gormLoanRepository) Issue(copyID, userID uint, borrowedAt, dueAt time.Time, maxActiveLoans int) (*model.Loan, error) {
	var loan model.Loan
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. var cp model.Copy. Load it: tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cp, copyID).
		//    If errors.Is(err, gorm.ErrRecordNotFound), return ErrInvalidReference. For any other error return it.
		var cp model.Copy
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cp, copyID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidReference
		}
		if err != nil {
			return err
		}

		// 2. If cp.Status != model.CopyStatusAvailable, return ErrCopyNotAvailable.
		if cp.Status != model.CopyStatusAvailable {
			return ErrCopyNotAvailable
		}

		// 3. var u model.User. Load it with the same lock: First(&u, userID). Not found -> ErrInvalidReference.
		var u model.User
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, userID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidReference
		}
		if err != nil {
			return err
		}

		// 4. var active int64. Count: tx.Model(&model.Loan{}).Where("user_id = ? AND returned_at IS NULL", userID).Count(&active).
		//    If int(active) >= maxActiveLoans, return ErrLoanLimitReached.
		var active int64
		err = tx.Model(&model.Loan{}).Where("user_id = ? AND returned_at IS NULL", userID).Count(&active).Error
		if err != nil {
			return err
		}
		if int(active) >= maxActiveLoans {
			return ErrLoanLimitReached
		}

		// 5. loan = model.Loan{CopyID: copyID, UserID: userID, BorrowedAt: borrowedAt, DueAt: dueAt}.
		//    Insert it with tx.Create(&loan). Return the error if any.
		loan = model.Loan{
			CopyID:     copyID,
			UserID:     userID,
			BorrowedAt: borrowedAt,
			DueAt:      dueAt,
		}
		err = tx.Create(&loan).Error
		if err != nil {
			return err
		}

		// 6. Update the copy: tx.Model(&model.Copy{}).Where("id = ?", copyID).Updates(map[string]interface{}{"status": "borrowed", "updated_at": time.Now()}). Return the error if any.
		err = tx.Model(&model.Copy{}).Where("id = ?", copyID).Updates(map[string]interface{}{"status": "borrowed", "updated_at": time.Now()}).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, translateLoanError(err)
	}
	// 7. Reload: var out model.Loan; err = r.db.Preload("Copy.Book").First(&out, loan.ID).Error.
	//    If err != nil return nil, translateLoanError(err). Otherwise return &out, nil.
	var out model.Loan
	err = r.db.Preload("Copy.Book").First(&out, loan.ID).Error
	if err != nil {
		return nil, translateLoanError(err)
	}
	return &out, nil
}
