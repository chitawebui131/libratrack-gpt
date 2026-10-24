package repository

import (
	"gorm.io/gorm"

	"libratrack/internal/model"
)

// GetBook loads one book with its authors, category and copies.
func (r *gormCatalogRepository) GetBook(id uint) (*model.Book, error) {
	// 1. var b model.Book.
	// 2. err := r.db.Preload("Authors").Preload("Category").Preload("Copies").First(&b, id).Error. If err != nil, return nil, translateError(err).
	// 3. Return &b, nil.
	var b model.Book
	err := r.db.Preload("Authors").Preload("Category").Preload("Copies").First(&b, id).Error
	if err != nil {
		return nil, translateError(err)
	}
	return &b, nil
}

type availabilityRow struct {
	BookID    uint `gorm:"column:book_id"`
	Total     int  `gorm:"column:total"`
	Available int  `gorm:"column:available"`
}

// AvailabilityForBooks returns the copy counts of all given books with ONE grouped query.
// Books without copies are present in the map with zeros. An empty input returns an empty map without a query.
func (r *gormCatalogRepository) AvailabilityForBooks(bookIDs []uint) (map[uint]model.CatalogAvailability, error) {
	// 1. out := map[uint]model.CatalogAvailability{}. If len(bookIDs) == 0, return out, nil.
	// 2. For every id in bookIDs set out[id] = model.CatalogAvailability{}.
	// 3. var rows []availabilityRow.
	// 4. err := r.db.Raw("SELECT book_id, COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 'available') AS available FROM copies WHERE book_id IN ? GROUP BY book_id", bookIDs).Scan(&rows).Error. If err != nil, return nil, translateError(err).
	// 5. For every row set out[row.BookID] = model.CatalogAvailability{Total: row.Total, Available: row.Available}.
	// 6. Return out, nil.
	out := map[uint]model.CatalogAvailability{}
	if len(bookIDs) == 0 {
		return out, nil
	}
	for _, id := range bookIDs {
		out[id] = model.CatalogAvailability{}
	}
	var rows []availabilityRow
	err := r.db.Raw("SELECT book_id, COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 'available') AS available FROM copies WHERE book_id IN ? GROUP BY book_id", bookIDs).Scan(&rows).Error
	if err != nil {
		return nil, translateError(err)
	}
	for _, row := range rows {
		out[row.BookID] = model.CatalogAvailability{Total: row.Total, Available: row.Available}
	}
	return out, nil
}

func NewCatalogRepository(db *gorm.DB) CatalogRepository {
	// Return &gormCatalogRepository{db: db}.
	return &gormCatalogRepository{db: db}
}
