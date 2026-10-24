package service

import (
	_ "embed"
	"encoding/csv"
	"strings"
	"sync"
)

//go:embed bbk_categories.csv
var bbkCSV string

var (
	bbkOnce  sync.Once
	bbkTable map[string]string
)

func loadBBK() {
	// 1. bbkTable = map[string]string{}.
	bbkTable = map[string]string{}
	// 2. records, err := csv.NewReader(strings.NewReader(bbkCSV)).ReadAll(). If err != nil, return.
	records, err := csv.NewReader(strings.NewReader(bbkCSV)).ReadAll()
	if err != nil {
		return
	}
	// 3. For every record at index >= 1 with len(record) >= 2: bbkTable[strings.TrimSpace(record[0])] = strings.TrimSpace(record[1]).
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) >= 2 {
			bbkTable[strings.TrimSpace(record[0])] = strings.TrimSpace(record[1])
		}
	}
}

func bbkLookup(key string) string {
	// bbkOnce.Do(loadBBK); return bbkTable[key].
	bbkOnce.Do(loadBBK)
	return bbkTable[key]
}

// UnknownAuthorName is the author used for books without an author.
func UnknownAuthorName() string {
	// Return bbkLookup("_unknown_author").
	return bbkLookup("_unknown_author")
}

func defaultCategoryName() string {
	// Return bbkLookup("_default_category").
	return bbkLookup("_default_category")
}

// CategoryFromBBK returns the category name for a BBK code like "34.61".
// It takes the leading digits (up to the first non-digit) and tries the longest prefix first.
func CategoryFromBBK(bbk string) string {
	// 1. digits := the leading run of ASCII digits of strings.TrimSpace(bbk).
	digits := strings.TrimLeft(strings.TrimSpace(bbk), "0123456789")
	// 2. For n := len(digits); n >= 1; n-- : if name := bbkLookup(digits[:n]); name != "" { return name }.
	for n := len(digits); n >= 1; n-- {
		if name := bbkLookup(digits[:n]); name != "" {
			return name
		}
	}
	// 3. Return defaultCategoryName().
	return defaultCategoryName()
}
