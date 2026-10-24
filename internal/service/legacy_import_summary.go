package service

import (
	"sort"
	"strings"

	"libratrack/internal/model"
)

// messageCode maps a report message to a short ASCII code. The template shows the Ukrainian text of the code.
func messageCode(msg string) string {
	switch {
	case msg == "name is empty":
		return "name_empty"
	case strings.HasPrefix(msg, "name is longer"):
		return "name_long"
	case strings.HasPrefix(msg, "countBooks must"):
		return "count_invalid"
	case msg == "inventar is empty":
		return "inventar_empty"
	case strings.HasPrefix(msg, "invalid inventar range"):
		return "inventar_range"
	case strings.HasPrefix(msg, "duplicate inventar number"):
		return "inventar_repeat_in_cell"
	case strings.Contains(msg, "is also used on line"):
		return "inventory_in_file_twice"
	case strings.HasSuffix(msg, " is too long"):
		return "field_too_long"
	case msg == "no author":
		return "no_author"
	case strings.HasPrefix(msg, "title may be cut"):
		return "title_cut"
	case strings.HasPrefix(msg, "ISBN ignored"):
		return "isbn_invalid"
	case strings.HasPrefix(msg, "language ignored"):
		return "language_unknown"
	case strings.HasPrefix(msg, "year ignored"):
		return "year_invalid"
	case strings.HasPrefix(msg, "inventar expanded"):
		return "inventar_expanded"
	case strings.HasPrefix(msg, "inventar has"):
		return "inventar_count_mismatch"
	case strings.Contains(msg, "conflict"):
		return "db_conflict"
	default:
		return "other"
	}
}

// Summarize groups the error and warning lines by (status, code).
// Lines with other statuses are ignored. Sort: errors first, then warnings; inside a status by Count descending, then by Code.
func Summarize(details []model.ImportRowResult) []model.ImportSummaryRow {
	index := map[string]int{}
	var rows []model.ImportSummaryRow

	for _, d := range details {
		if d.Status != "error" && d.Status != "warning" {
			continue
		}
		code := messageCode(d.Message)
		key := d.Status + "|" + code
		pos, exists := index[key]
		if !exists {
			rows = append(rows, model.ImportSummaryRow{
				Status:    d.Status,
				Code:      code,
				Count:     0,
				FirstLine: d.Line,
				Example:   d.Message,
			})
			pos = len(rows) - 1
			index[key] = pos
		}
		rows[pos].Count++
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Status != rows[j].Status {
			return rows[i].Status == "error" // errors first
		}
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count // higher count first
		}
		return rows[i].Code < rows[j].Code // alphabetical by code
	})

	return rows
}

// TrimDetails keeps at most limit detail lines (all error lines first, then the others in the original order)
// and stores the original number of lines in rep.DetailsTotal. limit <= 0 means no trimming (only DetailsTotal is set).
func TrimDetails(rep *model.ImportReport, limit int) {
	rep.DetailsTotal = len(rep.Details)
	if limit <= 0 || len(rep.Details) <= limit {
		return
	}

	out := make([]model.ImportRowResult, 0, limit)
	errorCount := 0

	// First pass: collect all error lines
	for _, d := range rep.Details {
		if d.Status == "error" && errorCount < limit {
			out = append(out, d)
			errorCount++
		}
	}

	// Second pass: collect remaining lines (non-errors) in original order
	for _, d := range rep.Details {
		if d.Status != "error" && len(out) < limit {
			out = append(out, d)
		}
	}

	rep.Details = out
}
