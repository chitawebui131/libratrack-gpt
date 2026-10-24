package repository

import "gorm.io/gorm"

// setAuthorPositions stores the order of the authors of a book.
// The first author gets position 0 (the main author), the next one 1, and so on.
// The rows of book_authors must already exist (after Association("Authors").Replace).
func setAuthorPositions(tx *gorm.DB, bookID uint, authorIDs []uint) error {
	for i, id := range authorIDs {
		err := tx.Exec("UPDATE book_authors SET position = ? WHERE book_id = ? AND author_id = ?", i, bookID, id).Error
		if err != nil {
			return err
		}
	}
	return nil
}
