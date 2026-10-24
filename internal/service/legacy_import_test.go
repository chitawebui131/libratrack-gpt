package service

import (
	"errors"
	"testing"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type fakeImportRepo struct {
	repository.ImportRepository
	calls  []model.ImportBook
	failOn string
}

func (f *fakeImportRepo) UpsertBook(b model.ImportBook) (model.ImportResult, error) {
	f.calls = append(f.calls, b)
	if b.Title == f.failOn {
		return model.ImportResult{}, errors.New("db error")
	}
	return model.ImportResult{BookCreated: true, CopiesAdded: len(b.Inventory)}, nil
}

// Helper function uses existing keys from legacy_columns.go
func rawRow(line int, code, inventar, name, count string) RawRow {
	return RawRow{
		Line: line,
		Cells: map[string]string{
			colCode:       code,
			colInventar:   inventar,
			colName:       name,
			colCountBooks: count,
		},
	}
}

func TestImportDryRunWritesNothing(t *testing.T) {
	repo := &fakeImportRepo{}
	service := NewImportService(repo)

	rows := []RawRow{
		rawRow(1, "A", "100", "Book A", "1"),
		rawRow(2, "B", "200", "Book B", "1"),
	}

	report := service.Run(rows, true)

	if len(repo.calls) != 0 {
		t.Errorf("Expected no calls to repository, got %d", len(repo.calls))
	}

	if report.Errors != 0 {
		t.Errorf("Expected 0 errors, got %d", report.Errors)
	}

	if report.Rows != 2 {
		t.Errorf("Expected 2 rows, got %d", report.Rows)
	}
}

func TestImportContinuesAfterError(t *testing.T) {
	repo := &fakeImportRepo{}
	service := NewImportService(repo)

	rows := []RawRow{
		rawRow(1, "A", "100", "Book A", "1"),
		rawRow(2, "B", "200", "", "1"), // Empty name -> error
		rawRow(3, "C", "300", "Book C", "1"),
	}

	report := service.Run(rows, false)

	if len(repo.calls) != 2 {
		t.Errorf("Expected 2 calls to repository, got %d", len(repo.calls))
	}

	if report.BooksCreated != 2 {
		t.Errorf("Expected 2 books created, got %d", report.BooksCreated)
	}

	if report.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", report.Errors)
	}
}

func TestImportDuplicateInventoryInFile(t *testing.T) {
	repo := &fakeImportRepo{}
	service := NewImportService(repo)

	rows := []RawRow{
		rawRow(1, "A", "100", "Book A", "1"),
		rawRow(2, "B", "100", "Book B", "1"), // Duplicate inventory
	}

	report := service.Run(rows, false)

	if len(repo.calls) != 1 {
		t.Errorf("Expected 1 call to repository, got %d", len(repo.calls))
	}

	if report.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", report.Errors)
	}
}

func TestImportRepositoryErrorIsReported(t *testing.T) {
	repo := &fakeImportRepo{failOn: "Bad"}
	service := NewImportService(repo)

	rows := []RawRow{
		rawRow(1, "A", "100", "Good", "1"),
		rawRow(2, "B", "200", "Bad", "1"), // This causes a db error but shouldn't block the next row
		rawRow(3, "C", "300", "Also Good", "1"),
	}

	report := service.Run(rows, false)

	// Since one bad row does NOT block others, all 3 rows must be processed by the loop
	if len(repo.calls) != 3 {
		t.Errorf("Expected 3 calls to repository, got %d", len(repo.calls))
	}

	if report.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", report.Errors)
	}
}
