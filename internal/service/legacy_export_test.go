package service

import (
	"bytes"
	"strings"
	"testing"

	"libratrack/internal/model"
)

func TestCompactInventory(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "numeric range",
			input:    []string{"22284", "22282", "22283"},
			expected: "22282-22284",
		},
		{
			name:     "non-numeric",
			input:    []string{"5", "7"},
			expected: "5; 7",
		},
		{
			name:     "non-numeric strings",
			input:    []string{"A-1", "A-2"},
			expected: "A-1; A-2",
		},
		{
			name:     "single element",
			input:    []string{"9"},
			expected: "9",
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compactInventory(tt.input)
			if result != tt.expected {
				t.Errorf("compactInventory(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWriteCSVHasBOMAndProtectsFormulas(t *testing.T) {
	rows := []model.ExportRow{
		{
			Title:     "=1+1",
			Inventory: []string{"1"},
			Count:     1,
		},
	}

	var buf bytes.Buffer
	err := WriteCSV(&buf, rows)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	output := buf.String()
	expectedBOM := "\xef\xbb\xbf"
	if !strings.HasPrefix(output, expectedBOM) {
		t.Errorf("Output does not start with BOM")
	}

	if !strings.Contains(output, "'=1+1") {
		t.Errorf("Output does not contain protected formula '=1+1'")
	}
	if strings.Contains(output, ";=1+1") {
		t.Errorf("Output contains unprotected formula ';=1+1'")
	}
}

func TestWriteCSVSkipsBooksWithoutCopies(t *testing.T) {
	rows := []model.ExportRow{
		{
			Title:     "Test Book",
			Inventory: nil,
			Count:     0,
		},
	}
	var buf bytes.Buffer
	err := WriteCSV(&buf, rows)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected only header line, got %d lines", len(lines))
	}
}

func TestExportThenImportRoundTrip(t *testing.T) {
	rows := []model.ExportRow{
		{
			Title:      "Book One",
			Authors:    "Ann Author; Bob Second",
			Inventory:  []string{"22282", "22283", "22284"},
			ISBN:       "9780306406157",
			UDK:        "821",
			Language:   "ukr",
			Department: "3",
			Count:      3,
		},
		{
			Title:     "Book Two",
			Authors:   "",
			Inventory: []string{"7"},
			Count:     1,
		},
	}

	var buf bytes.Buffer
	err := WriteXLSX(&buf, rows)
	if err != nil {
		t.Fatalf("WriteXLSX failed: %v", err)
	}

	// Read directly from the byte buffer. Avoid using excelize.OpenReader here.
	rowsBack, err := ReadXLSX(&buf)
	if err != nil {
		t.Fatalf("ReadXLSX failed: %v", err)
	}

	if len(rowsBack) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rowsBack))
	}

	// Convert raw rows using your rowToImportBook parser to validate internal values
	book1, _, err := rowToImportBook(rowsBack[0].Cells)
	if err != nil {
		t.Fatalf("rowToImportBook for Book One failed: %v", err)
	}

	if len(book1.Inventory) != 3 {
		t.Errorf("Book One should have 3 inventory numbers, got %d", len(book1.Inventory))
	}
	if book1.Inventory[0] != "22282" || book1.Inventory[1] != "22283" || book1.Inventory[2] != "22284" {
		t.Errorf("Book One inventory numbers incorrect")
	}
	if book1.ISBN != "9780306406157" {
		t.Errorf("Book One ISBN incorrect, got '%s'", book1.ISBN)
	}
	if book1.UDK != "821" {
		t.Errorf("Book One UDK incorrect")
	}
	if book1.Language != "ukr" {
		t.Errorf("Book One language incorrect")
	}
	if strings.Join(book1.Authors, "; ") != "Ann Author; Bob Second" {
		t.Errorf("Book One authors incorrect")
	}

	book2, _, err := rowToImportBook(rowsBack[1].Cells)
	if err != nil {
		t.Fatalf("rowToImportBook for Book Two failed: %v", err)
	}

	if len(book2.Inventory) != 1 {
		t.Errorf("Book Two should have 1 inventory number, got %d", len(book2.Inventory))
	}
	if book2.Inventory[0] != "7" {
		t.Errorf("Book Two inventory number incorrect")
	}
	if len(book2.Authors) != 1 || book2.Authors[0] != UnknownAuthorName() {
		t.Errorf("Book Two authors should be unknown author placeholder")
	}
}
