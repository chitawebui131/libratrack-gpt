package repository

import (
	"errors"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

// upsertBookRow creates the book or updates the existing one. It returns the book and true when it was created.
func upsertBookRow(tx *gorm.DB, b model.ImportBook, categoryID uint) (*model.Book, bool, error) {
	// 1. var book model.Book. var err error.
	var book model.Book
	var err error

	// 2. Find the book:
	//      if b.LegacyCode > 0: err = tx.Where("legacy_code = ?", b.LegacyCode).First(&book).Error
	//      else if b.ISBN != "": err = tx.Where("isbn = ?", b.ISBN).First(&book).Error
	//      else: err = tx.Where("lower(title) = lower(?) AND published_year = ?", b.Title, b.Year).First(&book).Error
	if b.LegacyCode > 0 {
		err = tx.Where("legacy_code = ?", b.LegacyCode).First(&book).Error
	} else if b.ISBN != "" {
		err = tx.Where("isbn = ?", b.ISBN).First(&book).Error
	} else {
		err = tx.Where("lower(title) = lower(?) AND published_year = ?", b.Title, b.Year).First(&book).Error
	}

	// 3. If err != nil and !errors.Is(err, gorm.ErrRecordNotFound), return nil, false, err.
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	// 4. If errors.Is(err, gorm.ErrRecordNotFound):
	//      book = model.Book{Title: b.Title, CategoryID: categoryID, PublishedYear: b.Year, Publisher: b.Publisher, PublicationCity: b.City, PageCount: b.Pages, BBK: b.BBK, UDK: b.UDK, Language: b.Language, ISBN: b.ISBN}.
	//      If b.LegacyCode > 0 { code := b.LegacyCode; book.LegacyCode = &code }.
	//      If err := tx.Create(&book).Error; err != nil, return nil, false, err. Return &book, true, nil.
	if errors.Is(err, gorm.ErrRecordNotFound) {
		book = model.Book{
			Title:           b.Title,
			CategoryID:      categoryID,
			PublishedYear:   b.Year,
			Publisher:       b.Publisher,
			PublicationCity: b.City,
			PageCount:       b.Pages,
			BBK:             b.BBK,
			UDK:             b.UDK,
			Language:        b.Language,
			ISBN:            b.ISBN,
		}
		if b.LegacyCode > 0 {
			code := b.LegacyCode
			book.LegacyCode = &code
		}
		if err := tx.Create(&book).Error; err != nil {
			return nil, false, err
		}
		return &book, true, nil
	}

	// 5. Update the existing book: values := map[string]interface{}{"title": b.Title, "category_id": categoryID, "published_year": b.Year, "publisher": b.Publisher, "publication_city": b.City, "page_count": b.Pages, "bbk": b.BBK}.
	//      If b.ISBN != "" { values["isbn"] = b.ISBN }. If b.UDK != "" { values["udk"] = b.UDK }. If b.Language != "" { values["language"] = b.Language }.
	//      tx.Model(&model.Book{}).Where("id = ?", book.ID).Updates(values). Return the error if any.
	values := map[string]interface{}{
		"title":            b.Title,
		"category_id":      categoryID,
		"published_year":   b.Year,
		"publisher":        b.Publisher,
		"publication_city": b.City,
		"page_count":       b.Pages,
		"bbk":              b.BBK,
	}
	if b.ISBN != "" {
		values["isbn"] = b.ISBN
	}
	if b.UDK != "" {
		values["udk"] = b.UDK
	}
	if b.Language != "" {
		values["language"] = b.Language
	}
	err = tx.Model(&model.Book{}).Where("id = ?", book.ID).Updates(values).Error
	if err != nil {
		return nil, false, err
	}

	// 6. Return &book, false, nil.
	return &book, false, nil
}
