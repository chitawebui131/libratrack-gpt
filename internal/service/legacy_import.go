package service

import (
	"fmt"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type ImportService struct {
	repo repository.ImportRepository
}

func NewImportService(repo repository.ImportRepository) *ImportService {
	return &ImportService{repo: repo}
}

// Run imports the rows and returns a report.
func (s *ImportService) Run(rows []RawRow, dryRun bool) *model.ImportReport {
	// 1. rep := &model.ImportReport{DryRun: dryRun, Rows: len(rows)}. usedInventory := map[string]int{} (inventory number -> line).
	rep := &model.ImportReport{DryRun: dryRun, Rows: len(rows)}
	usedInventory := map[string]int{} // inventory number -> line

	// 2. For every row:
	for _, row := range rows {
		// book, warnings, err := rowToImportBook(row.Cells).
		book, warnings, err := rowToImportBook(row.Cells)
		// If err != nil: add a detail {Line: row.Line, Status: "error", Message: err.Error()}; rep.Errors++; continue.
		if err != nil {
			rep.Details = append(rep.Details, model.ImportRowResult{Line: row.Line, Status: "error", Message: err.Error()})
			rep.Errors++
			continue
		}
		// For every inventory number n of book.Inventory: if line, dup := usedInventory[n]; dup { error "inventory number n is also used on line X" -> rep.Errors++ and skip the whole row (continue the outer loop) }; usedInventory[n] = row.Line. (Check all numbers first, register them only when none is a duplicate.)
		duplicateFound := false
		for _, n := range book.Inventory {
			if line, dup := usedInventory[n]; dup {
				rep.Details = append(rep.Details, model.ImportRowResult{Line: row.Line, Status: "error", Message: fmt.Sprintf("inventory number %s is also used on line %d", n, line)})
				rep.Errors++
				duplicateFound = true
				break
			}
		}
		if duplicateFound {
			continue
		}
		for _, n := range book.Inventory {
			usedInventory[n] = row.Line
		}
		// For every warning w: add a detail {Line, "warning", w}; rep.Warnings++.
		for _, w := range warnings {
			rep.Details = append(rep.Details, model.ImportRowResult{Line: row.Line, Status: "warning", Message: w})
			rep.Warnings++
		}
		// If dryRun: continue.
		if dryRun {
			continue
		}
		// res, err := s.repo.UpsertBook(*book). If err != nil: add a detail {Line, "error", err.Error()}; rep.Errors++; continue.
		res, err := s.repo.UpsertBook(*book)
		if err != nil {
			rep.Details = append(rep.Details, model.ImportRowResult{Line: row.Line, Status: "error", Message: err.Error()})
			rep.Errors++
			continue
		}
		// If res.BookCreated { rep.BooksCreated++; status "created" } else { rep.BooksUpdated++; status "updated" }. rep.CopiesAdded += res.CopiesAdded; rep.CopiesSkipped += res.CopiesSkipped.
		var status string
		if res.BookCreated {
			rep.BooksCreated++
			status = "created"
		} else {
			rep.BooksUpdated++
			status = "updated"
		}
		rep.CopiesAdded += res.CopiesAdded
		rep.CopiesSkipped += res.CopiesSkipped
		// Add a detail {Line, status, fmt.Sprintf("%s (copies added: %d, skipped: %d)", book.Title, res.CopiesAdded, res.CopiesSkipped)}.
		rep.Details = append(rep.Details, model.ImportRowResult{Line: row.Line, Status: status, Message: fmt.Sprintf("%s (copies added: %d, skipped: %d)", book.Title, res.CopiesAdded, res.CopiesSkipped)})
	}
	// 3. Return rep.
	rep.Summary = Summarize(rep.Details)
	return rep
}
