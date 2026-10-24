package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

// SaveBook creates the book (b.ID == 0) or updates it, and replaces its authors. Everything is ONE transaction.
func (r *gormPanelRepository) SaveBook(b *model.Book, authorIDs []uint) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. var authors []model.Author. If len(authorIDs) > 0: tx.Find(&authors, authorIDs). If len(authors) != number of unique ids in authorIDs, return ErrInvalidReference.
		var authors []model.Author
		if len(authorIDs) > 0 {
			if err := tx.Find(&authors, authorIDs).Error; err != nil {
				return err
			}
			// Build uniqueIDsSlice: the ids of authorIDs without duplicates, keeping the original order.
			uniqueIDsSlice := uniqueIDs(authorIDs)
			if len(authors) != len(uniqueIDsSlice) {
				return ErrInvalidReference
			}
		}

		// 2. If b.ID == 0: if err := tx.Create(b).Error; err != nil, return err.
		//    Otherwise: res := tx.Model(&model.Book{}).Where("id = ?", b.ID).Select("Title", "ISBN", "CategoryID", "Publisher", "PublishedYear", "Description").Updates(b). If res.Error != nil return it. If res.RowsAffected == 0 return ErrNotFound.
		if b.ID == 0 {
			if err := tx.Create(b).Error; err != nil {
				return err
			}
		} else {
			res := tx.Model(&model.Book{}).Where("id = ?", b.ID).Select("Title", "ISBN", "CategoryID", "Publisher", "PublishedYear", "Description", "UDK", "Language", "PublicationCity", "PageCount", "BBK").Updates(b)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return ErrNotFound
			}
		}

		// 3. return tx.Model(&model.Book{ID: b.ID}).Association("Authors").Replace(&authors).
		err := tx.Model(&model.Book{ID: b.ID}).Association("Authors").Replace(&authors)
		if err != nil {
			return err
		}

		// 4. After Association("Authors").Replace(&authors) succeeds, call setAuthorPositions(tx, b.ID, uniqueIDsSlice) and return its error if any.
		return setAuthorPositions(tx, b.ID, uniqueIDs(authorIDs))
	})
	// 5. If err != nil return translateLoanError(err). Otherwise nil.
	if err != nil {
		return translateLoanError(err)
	}
	return nil
}

// DeleteBook removes a book with its copies and author links. A book whose copies have loan history cannot be removed (ErrInvalidReference).
func (r *gormPanelRepository) DeleteBook(id uint) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. var n int64. tx.Model(&model.Book{}).Where("id = ?", id).Count(&n). If n == 0 return ErrNotFound.
		var n int64
		tx.Model(&model.Book{}).Where("id = ?", id).Count(&n)
		if n == 0 {
			return ErrNotFound
		}
		// 2. var loans int64. tx.Table("loans").Joins("JOIN copies c ON c.id = loans.copy_id").Where("c.book_id = ?", id).Count(&loans). If loans > 0 return ErrInvalidReference.
		var loans int64
		tx.Table("loans").Joins("JOIN copies c ON c.id = loans.copy_id").Where("c.book_id = ?", id).Count(&loans)
		if loans > 0 {
			return ErrInvalidReference
		}
		// 3. tx.Where("book_id = ?", id).Delete(&model.Copy{}). Return the error if any.
		err := tx.Where("book_id = ?", id).Delete(&model.Copy{})
		if err.Error != nil {
			return err.Error
		}
		// 4. tx.Exec("DELETE FROM book_authors WHERE book_id = ?", id). Return the error if any.
		err = tx.Exec("DELETE FROM book_authors WHERE book_id = ?", id)
		if err.Error != nil {
			return err.Error
		}
		// 5. res := tx.Delete(&model.Book{}, id). If res.Error != nil return it. If res.RowsAffected == 0 return ErrNotFound.
		res := tx.Delete(&model.Book{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		// 6. Return nil.
		return nil
	})
	return translateLoanError(err)
}

func NewPanelRepository(db *gorm.DB) PanelRepository {
	// Return &gormPanelRepository{db: db}.
	return &gormPanelRepository{db: db}
}
