package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

type ImportRepository interface {
	UpsertBook(b model.ImportBook) (model.ImportResult, error)
}

type gormImportRepository struct {
	db *gorm.DB
}

func NewImportRepository(db *gorm.DB) ImportRepository {
	return &gormImportRepository{db: db}
}

func (r *gormImportRepository) UpsertBook(b model.ImportBook) (model.ImportResult, error) {
	var res model.ImportResult
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. categoryID, err := findOrCreateCategory(tx, b.CategoryName). Return the error if any.
		categoryID, err := findOrCreateCategory(tx, b.CategoryName)
		if err != nil {
			return err
		}
		// 2. authors, err := findOrCreateAuthors(tx, b.Authors). Return the error if any.
		authors, err := findOrCreateAuthors(tx, b.Authors)
		if err != nil {
			return err
		}
		// 3. book, created, err := upsertBookRow(tx, b, categoryID). Return the error if any. res.BookCreated = created.
		book, created, err := upsertBookRow(tx, b, categoryID)
		if err != nil {
			return err
		}
		res.BookCreated = created
		// 4. Replace the authors: tx.Model(book).Association("Authors").Replace(authors). Return the error if any. Then ids := the ids of authors in the same order; if err := setAuthorPositions(tx, book.ID, ids); err != nil return err.
		err = tx.Model(book).Association("Authors").Replace(authors)
		if err != nil {
			return err
		}
		ids := make([]uint, len(authors))
		for i, a := range authors {
			ids[i] = a.ID
		}
		if err := setAuthorPositions(tx, book.ID, ids); err != nil {
			return err
		}
		// 5. added, skipped, err := addCopies(tx, book.ID, b.Inventory, b.Department). Return the error if any. res.CopiesAdded = added; res.CopiesSkipped = skipped.
		added, skipped, err := addCopies(tx, book.ID, b.Inventory, b.Department)
		if err != nil {
			return err
		}
		res.CopiesAdded = added
		res.CopiesSkipped = skipped
		return nil
	})
	if err != nil {
		return model.ImportResult{}, translateLoanError(err)
	}
	return res, nil
}
