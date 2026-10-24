package repository

import (
	"errors"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

func addCopies(tx *gorm.DB, bookID uint, inventory []string, department string) (added, skipped int, err error) {
	// 1. For every number in inventory:
	//      var cp model.Copy. e := tx.Where("inventory_no = ?", number).First(&cp).Error.
	//      If e == nil: if cp.BookID == bookID { skipped++; continue }; return 0, 0, ErrConflict.
	//      If !errors.Is(e, gorm.ErrRecordNotFound): return 0, 0, e.
	//      cp = model.Copy{BookID: bookID, InventoryNo: number, Status: model.CopyStatusAvailable, Department: department}.
	//      If e := tx.Create(&cp).Error; e != nil: return 0, 0, e.
	//      added++.
	for _, number := range inventory {
		var cp model.Copy
		e := tx.Where("inventory_no = ?", number).First(&cp).Error
		if e == nil {
			if cp.BookID == bookID {
				skipped++
				continue
			}
			return 0, 0, ErrConflict
		}
		if !errors.Is(e, gorm.ErrRecordNotFound) {
			return 0, 0, e
		}
		cp = model.Copy{BookID: bookID, InventoryNo: number, Status: model.CopyStatusAvailable, Department: department}
		e = tx.Create(&cp).Error
		if e != nil {
			return 0, 0, e
		}
		added++
	}
	// 2. Return added, skipped, nil.
	return added, skipped, nil
}
