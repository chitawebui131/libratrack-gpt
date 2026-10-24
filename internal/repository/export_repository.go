package repository

import (
	"strings"

	"gorm.io/gorm"

	"libratrack/internal/model"
)

type ExportRepository interface {
	ListExportRows() ([]model.ExportRow, error)
}

type gormExportRepository struct {
	db *gorm.DB
}

func NewExportRepository(db *gorm.DB) ExportRepository {
	return &gormExportRepository{db: db}
}

type exportScan struct {
	LegacyCode *int   `gorm:"column:legacy_code"`
	Title      string `gorm:"column:title"`
	Authors    string `gorm:"column:authors"`
	Year       int    `gorm:"column:year"`
	City       string `gorm:"column:city"`
	Publisher  string `gorm:"column:publisher"`
	Department string `gorm:"column:department"`
	Pages      int    `gorm:"column:pages"`
	BBK        string `gorm:"column:bbk"`
	UDK        string `gorm:"column:udk"`
	ISBN       string `gorm:"column:isbn"`
	Language   string `gorm:"column:language"`
	Inventory  string `gorm:"column:inventory"`
	CopyCount  int    `gorm:"column:copy_count"`
}

const exportSQL = `SELECT b.legacy_code, b.title,
    COALESCE((SELECT string_agg(a.full_name, '; ' ORDER BY ba.position, a.full_name) FROM book_authors ba JOIN authors a ON a.id = ba.author_id WHERE ba.book_id = b.id), '') AS authors,
    COALESCE(b.published_year, 0) AS year, b.publication_city AS city, b.publisher, c.department AS department,
    b.page_count AS pages, b.bbk, b.udk, b.isbn, b.language,
    string_agg(c.inventory_no, E'\n' ORDER BY c.inventory_no) AS inventory,
    COUNT(*) AS copy_count
FROM books b
JOIN copies c ON c.book_id = b.id
GROUP BY b.id, c.department
ORDER BY b.title, b.id, c.department`

func (r *gormExportRepository) ListExportRows() ([]model.ExportRow, error) {
	// 1. var scans []exportScan. err := r.db.Raw(exportSQL).Scan(&scans).Error. If err != nil, return nil, translateError(err).
	var scans []exportScan
	err := r.db.Raw(exportSQL).Scan(&scans).Error
	if err != nil {
		return nil, translateError(err)
	}
	// 2. out := make([]model.ExportRow, 0, len(scans)).
	out := make([]model.ExportRow, 0, len(scans))
	// 3. For every s in scans: inv := []string{}; if s.Inventory != "" { inv = strings.Split(s.Inventory, "\n") }. Append model.ExportRow{LegacyCode: s.LegacyCode, Title: s.Title, Authors: s.Authors, Year: s.Year, City: s.City, Publisher: s.Publisher, Department: s.Department, Pages: s.Pages, BBK: s.BBK, UDK: s.UDK, ISBN: s.ISBN, Language: s.Language, Inventory: inv, Count: s.CopyCount}.
	for _, s := range scans {
		inv := []string{}
		if s.Inventory != "" {
			inv = strings.Split(s.Inventory, "\n")
		}
		out = append(out, model.ExportRow{
			LegacyCode: s.LegacyCode,
			Title:      s.Title,
			Authors:    s.Authors,
			Year:       s.Year,
			City:       s.City,
			Publisher:  s.Publisher,
			Department: s.Department,
			Pages:      s.Pages,
			BBK:        s.BBK,
			UDK:        s.UDK,
			ISBN:       s.ISBN,
			Language:   s.Language,
			Inventory:  inv,
			Count:      s.CopyCount,
		})
	}
	// 4. Return out, nil.
	return out, nil
}
