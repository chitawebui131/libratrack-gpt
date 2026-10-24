package service

import (
	"bytes"
	"os"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestReadSampleCSV(t *testing.T) {
	// 1. f, err := os.Open("testdata/books_sample.csv"); fatal on error; defer f.Close().
	f, err := os.Open("testdata/books_sample.csv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	// 2. rows, err := ReadCSV(f). Fatal on error. There must be exactly 11 rows.
	rows, err := ReadCSV(f)
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(rows) != 11 {
		t.Fatalf("expected 11 rows, got %d", len(rows))
	}

	// 3. total := 0; authorless := 0. For every row: book, warnings, err := rowToImportBook(row.Cells). Fail the test if err != nil. total += len(book.Inventory). If len(book.Authors) == 1 && book.Authors[0] == UnknownAuthorName() { authorless++ }. Make sure book.Title is not empty and book.CategoryName is not empty.
	total := 0
	authorless := 0
	var targetBook *RawRow

	for _, row := range rows {
		book, w, err := rowToImportBook(row.Cells)
		if err != nil {
			t.Fatalf("failed to convert row at line %d: %v", row.Line, err)
		}
		_ = w // warnings ignored for this test

		total += len(book.Inventory)
		if len(book.Authors) == 1 && book.Authors[0] == UnknownAuthorName() {
			authorless++
		}

		if book.Title == "" || book.CategoryName == "" {
			t.Errorf("empty title or category name at line %d", row.Line)
		}

		if book.LegacyCode == 35 {
			// Capture the specific row with code 35 for deeper validation
			targetRow := row
			targetBook = &targetRow
		}
	}

	// 4. total must be 27 (ten rows with one copy and one row with 17 copies). authorless must be 3.
	if total != 27 {
		t.Errorf("expected 27 total copies, got %d", total)
	}

	if authorless != 3 {
		t.Errorf("expected 3 authorless books, got %d", authorless)
	}

	// 5. The row with LegacyCode 35 has 17 inventory numbers, the first "22282" and the last "22298".
	if targetBook == nil {
		t.Fatalf("book with legacy code 35 not found in file")
	}

	// Validate the specific expanded range for code 35
	book35, _, _ := rowToImportBook(targetBook.Cells)
	if len(book35.Inventory) != 17 {
		t.Errorf("expected 17 inventory numbers for code 35, got %d", len(book35.Inventory))
	} else {
		if book35.Inventory[0] != "22282" {
			t.Errorf("expected first inventory number '22282', got '%s'", book35.Inventory[0])
		}
		if book35.Inventory[16] != "22298" {
			t.Errorf("expected last inventory number '22298', got '%s'", book35.Inventory[16])
		}
	}
}

func TestReadCSVCommaDelimiter(t *testing.T) {
	// ASCII CSV with "," as delimiter and a BOM prefix: "\ufeffcode,inventar,name,countBooks\n1,100,Title,1\n". ReadCSV must return 1 row with Cells[colName] == "Title".
	input := "\ufeffcode,inventar,name,countBooks\n1,100,Title,1\n"
	rows, err := ReadCSV(bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].Cells[colName] != "Title" {
		t.Errorf("expected Title, got '%s'", rows[0].Cells[colName])
	}
}

func TestReadXLSXRoundTrip(t *testing.T) {
	// 1. Build a workbook in memory with excelize.NewFile(): row 1 = "code","inventar","name","author","countBooks"; row 2 = 7, "300-302", "Title", "Author", 3 (use f.SetSheetRow).
	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Sheet1"

	header := []interface{}{"code", "inventar", "name", "author", "countBooks"}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		t.Fatalf("failed to set header: %v", err)
	}

	rowData := []interface{}{7, "300-302", "Title", "Author", 3}
	if err := f.SetSheetRow(sheet, "A2", &rowData); err != nil {
		t.Fatalf("failed to set row data: %v", err)
	}

	// 2. Write it into a bytes.Buffer with f.Write(&buf).
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("failed to write workbook: %v", err)
	}

	// 3. rows, err := ReadXLSX(&buf). One row; Cells[colInventar] == "300-302"; Cells[colCountBooks] == "3".
	rows, err := ReadXLSX(&buf)
	if err != nil {
		t.Fatalf("failed to read XLSX: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].Cells[colInventar] != "300-302" {
		t.Errorf("expected '300-302', got '%s'", rows[0].Cells[colInventar])
	}
	if rows[0].Cells[colCountBooks] != "3" {
		t.Errorf("expected '3', got '%s'", rows[0].Cells[colCountBooks])
	}
}

func TestMissingRequiredColumn(t *testing.T) {
	// ReadCSV of "code,name\n1,Title\n" must return an error (inventar and countBooks are missing).
	input := "code,name\n1,Title\n"
	_, err := ReadCSV(bytes.NewReader([]byte(input)))
	if err == nil {
		t.Error("expected error due to missing required columns, got nil")
	}
}
