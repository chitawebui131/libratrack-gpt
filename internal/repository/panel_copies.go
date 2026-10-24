package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"libratrack/internal/model"
)

// AddCopy adds a copy with status available to a book.
func (r *gormPanelRepository) AddCopy(bookID uint, inventoryNo, department string) error {
	// 1. var count int64. r.db.Model(&model.Book{}).Where("id = ?", bookID).Count(&count). If count == 0, return ErrNotFound.
	var count int64
	r.db.Model(&model.Book{}).Where("id = ?", bookID).Count(&count)
	if count == 0 {
		return ErrNotFound
	}
	// 2. cp := model.Copy{BookID: bookID, InventoryNo: inventoryNo, Status: model.CopyStatusAvailable}.
	cp := model.Copy{BookID: bookID, InventoryNo: inventoryNo, Status: model.CopyStatusAvailable, Department: department}
	// 3. err := r.db.Create(&cp).Error. If err != nil return translateLoanError(err). Otherwise nil. (A duplicate inventory number becomes ErrConflict.)
	err := r.db.Create(&cp).Error
	if err != nil {
		return translateLoanError(err)
	}
	return nil
}

// UpdateCopyStatus changes the status manually. Allowed: available, lost, under_repair.
// A copy with an active loan is never changed.
func (r *gormPanelRepository) UpdateCopyStatus(copyID uint, status string) error {
	// 1. If status is not one of "available", "lost", "under_repair", return ErrInvalidReference.
	if status != model.CopyStatusAvailable && status != model.CopyStatusLost && status != model.CopyStatusUnderRepair {
		return ErrInvalidReference
	}
	// 2. err := r.db.Transaction(func(tx *gorm.DB) error { ... }) where the function does:
	err := r.db.Transaction(func(tx *gorm.DB) error {
		//      a) var cp model.Copy; load with tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cp, copyID). If errors.Is(err, gorm.ErrRecordNotFound) return ErrNotFound; other errors return err.
		var cp model.Copy
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cp, copyID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		//      b) var active int64; tx.Model(&model.Loan{}).Where("copy_id = ? AND returned_at IS NULL", copyID).Count(&active). If active > 0 return ErrCopyNotAvailable.
		var active int64
		tx.Model(&model.Loan{}).Where("copy_id = ? AND returned_at IS NULL", copyID).Count(&active)
		if active > 0 {
			return ErrCopyNotAvailable
		}
		//      c) return tx.Model(&model.Copy{}).Where("id = ?", copyID).Updates(map[string]interface{}{"status": status, "updated_at": time.Now()}).Error.
		return tx.Model(&model.Copy{}).Where("id = ?", copyID).Updates(map[string]interface{}{"status": status, "updated_at": time.Now()}).Error
	})
	// 3. If err != nil return translateLoanError(err). Otherwise nil.
	if err != nil {
		return translateLoanError(err)
	}
	return nil
}

// UpdateCopy changes the inventory number and the department. A duplicate inventory number gives ErrConflict.
// 1. res := r.db.Model(&model.Copy{}).Where("id = ?", copyID).Select("InventoryNo", "Department").Updates(&model.Copy{InventoryNo: inventoryNo, Department: department}). If res.Error != nil return translateLoanError(res.Error).
// 2. If res.RowsAffected == 0 return ErrNotFound. Otherwise nil.
func (r *gormPanelRepository) UpdateCopy(copyID uint, inventoryNo, department string) error {
	res := r.db.Model(&model.Copy{}).Where("id = ?", copyID).Select("InventoryNo", "Department").Updates(&model.Copy{InventoryNo: inventoryNo, Department: department})
	if res.Error != nil {
		return translateLoanError(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteCopy removes a copy that never had a loan. A copy with loan history gives ErrInvalidReference.
// 1. var n int64; r.db.Model(&model.Copy{}).Where("id = ?", copyID).Count(&n). If n == 0 return ErrNotFound.
// 2. var loans int64; r.db.Model(&model.Loan{}).Where("copy_id = ?", copyID).Count(&loans). If loans > 0 return ErrInvalidReference.
// 3. res := r.db.Delete(&model.Copy{}, copyID). If res.Error != nil return translateLoanError(res.Error). Otherwise nil.
func (r *gormPanelRepository) DeleteCopy(copyID uint) error {
	var n int64
	r.db.Model(&model.Copy{}).Where("id = ?", copyID).Count(&n)
	if n == 0 {
		return ErrNotFound
	}
	var loans int64
	r.db.Model(&model.Loan{}).Where("copy_id = ?", copyID).Count(&loans)
	if loans > 0 {
		return ErrInvalidReference
	}
	res := r.db.Delete(&model.Copy{}, copyID)
	if res.Error != nil {
		return translateLoanError(res.Error)
	}
	return nil
}
