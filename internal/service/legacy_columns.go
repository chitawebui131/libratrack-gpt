package service

import (
	"errors"
	"strings"
)

const maxImportRows = 20000

// Canonical column keys.
const (
	colCode       = "code"
	colInventar   = "inventar"
	colName       = "name"
	colAuthor     = "author"
	colYear       = "year"
	colCity       = "city"
	colPublishing = "publishing"
	colDepartment = "department"
	colCountPage  = "countpage"
	colBBK        = "bbk"
	colCountBooks = "countbooks"
	colUDK        = "udk"
	colISBN       = "isbn"
	colLanguage   = "language"
)

// legacyHeaders are the header cells written by the exporter, in the original order.
var legacyHeaders = []string{"\u041a\u043e\u0434", "inventar", "name", "author", "year", "city", "publishing", "department", "countPage", "BBK", "countBooks", "UDK", "ISBN", "language"}

// RawRow is one data row with canonical column keys. Line is the 1-based line in the file.
type RawRow struct {
	Line  int
	Cells map[string]string
}

// canonicalHeader maps a header cell to a canonical key. It returns "" for an unknown header.
func canonicalHeader(h string) string {
	s := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(h, "\ufeff")))
	switch s {
	case "code", "\u043a\u043e\u0434":
		return colCode
	case "inventar":
		return colInventar
	case "name":
		return colName
	case "author":
		return colAuthor
	case "year":
		return colYear
	case "city":
		return colCity
	case "publishing":
		return colPublishing
	case "department":
		return colDepartment
	case "countpage":
		return colCountPage
	case "bbk":
		return colBBK
	case "countbooks":
		return colCountBooks
	case "udk":
		return colUDK
	case "isbn":
		return colISBN
	case "language":
		return colLanguage
	default:
		return ""
	}
}

// rowsFromTable turns a table (first row = header) into RawRows.
func rowsFromTable(table [][]string) ([]RawRow, error) {
	if len(table) == 0 {
		return nil, errors.New("the file is empty")
	}
	keys := make([]string, len(table[0]))
	found := map[string]bool{}
	for i, h := range table[0] {
		keys[i] = canonicalHeader(h)
		if keys[i] != "" {
			found[keys[i]] = true
		}
	}
	for _, required := range []string{colInventar, colName, colCountBooks} {
		if !found[required] {
			return nil, errors.New("missing required column: " + required)
		}
	}
	if len(table)-1 > maxImportRows {
		return nil, errors.New("too many rows")
	}
	out := []RawRow{}
	for i, row := range table {
		if i == 0 {
			continue
		}
		cells := map[string]string{}
		empty := true
		for j, value := range row {
			if j >= len(keys) || keys[j] == "" {
				continue
			}
			v := strings.TrimSpace(value)
			cells[keys[j]] = v
			if v != "" {
				empty = false
			}
		}
		if empty {
			continue
		}
		out = append(out, RawRow{Line: i + 1, Cells: cells})
	}
	return out, nil
}
