package service

import (
	"sort"
	"strconv"
	"strings"

	"libratrack/internal/model"
)

func isNumeric(s string) bool {
	// Return true when s is not empty and every byte is an ASCII digit.
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// sortInventory sorts numbers numerically when both are numeric, otherwise as text.
func sortInventory(inv []string) []string {
	// 1. out := append([]string(nil), inv...).
	// 2. sort.Slice(out, func(i, j int) bool: if isNumeric(out[i]) && isNumeric(out[j]) { a, _ := strconv.Atoi(out[i]); b, _ := strconv.Atoi(out[j]); return a < b }; return out[i] < out[j]).
	// 3. Return out.
	out := append([]string(nil), inv...)
	sort.Slice(out, func(i, j int) bool {
		if isNumeric(out[i]) && isNumeric(out[j]) {
			a, _ := strconv.Atoi(out[i])
			b, _ := strconv.Atoi(out[j])
			return a < b
		}
		return out[i] < out[j]
	})
	return out
}

// compactInventory writes consecutive numbers as a range ("22282-22298"). Anything else is joined with "; ".
func compactInventory(inv []string) string {
	// 1. s := sortInventory(inv). If len(s) == 0 return "". If len(s) == 1 return s[0].
	// 2. If every element is numeric: check that s[i] == strconv.Itoa(first + i) for all i (first = the first number). If so return s[0] + "-" + s[len(s)-1].
	// 3. Otherwise return strings.Join(s, "; ").
	s := sortInventory(inv)
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return s[0]
	}
	// Check if all elements are numeric
	allNumeric := true
	for _, v := range s {
		if !isNumeric(v) {
			allNumeric = false
			break
		}
	}
	if allNumeric {
		first, _ := strconv.Atoi(s[0])
		isConsecutive := true
		for i, v := range s {
			current, _ := strconv.Atoi(v)
			if current != first+i {
				isConsecutive = false
				break
			}
		}
		if isConsecutive {
			last, _ := strconv.Atoi(s[len(s)-1])
			return strconv.Itoa(first) + "-" + strconv.Itoa(last)
		}
	}
	return strings.Join(s, "; ")
}

// exportCells returns the 14 cells of a row in the order of legacyHeaders.
func exportCells(row model.ExportRow) []string {
	// 1. code := "". If row.LegacyCode != nil { code = strconv.Itoa(*row.LegacyCode) }.
	// 2. authors := row.Authors. If authors == UnknownAuthorName() { authors = "" }.
	// 3. year := ""; if row.Year > 0 { year = strconv.Itoa(row.Year) }. pages := ""; if row.Pages > 0 { pages = strconv.Itoa(row.Pages) }.
	// 4. Return []string{code, compactInventory(row.Inventory), row.Title, authors, year, row.City, row.Publisher, row.Department, pages, row.BBK, strconv.Itoa(len(row.Inventory)), row.UDK, row.ISBN, row.Language}.
	code := ""
	if row.LegacyCode != nil {
		code = strconv.Itoa(*row.LegacyCode)
	}
	authors := row.Authors
	if authors == UnknownAuthorName() {
		authors = ""
	}
	year := ""
	if row.Year > 0 {
		year = strconv.Itoa(row.Year)
	}
	pages := ""
	if row.Pages > 0 {
		pages = strconv.Itoa(row.Pages)
	}
	return []string{code, compactInventory(row.Inventory), row.Title, authors, year, row.City, row.Publisher, row.Department, pages, row.BBK, strconv.Itoa(len(row.Inventory)), row.UDK, row.ISBN, row.Language}
}
