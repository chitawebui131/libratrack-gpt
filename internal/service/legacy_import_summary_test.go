package service

import (
	"testing"

	"libratrack/internal/model"
)

func TestMessageCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"name is empty", "name_empty"},
		{"inventory number 5 is also used on line 3", "inventory_in_file_twice"},
		{"no author", "no_author"},
		{"title may be cut in the source (50 characters)", "title_cut"},
		{"ISBN ignored: invalid", "isbn_invalid"},
		{"city is too long", "field_too_long"},
		{"duplicate inventar number: 7", "inventar_repeat_in_cell"},
		{"something unexpected", "other"},
	}

	for _, tt := range tests {
		actual := messageCode(tt.input)
		if actual != tt.expected {
			t.Errorf("messageCode(%q) = %q, want %q", tt.input, actual, tt.expected)
		}
	}
}

func TestSummarizeGroupsAndSorts(t *testing.T) {
	details := []model.ImportRowResult{
		{Status: "warning", Message: "no author", Line: 4},
		{Status: "warning", Message: "no author", Line: 5},
		{Status: "warning", Message: "no author", Line: 9},
		{Status: "error", Message: "name is empty", Line: 7},
		{Status: "error", Message: "inventory number 1 is also used on line 2", Line: 11},
		{Status: "error", Message: "inventory number 1 is also used on line 2", Line: 12},
		{Status: "created", Message: "something created", Line: 13},
	}

	result := Summarize(details)

	if len(result) != 3 {
		t.Fatalf("expected 3 summary rows, got %d", len(result))
	}

	// First row: inventory_in_file_twice error with Count 2 and FirstLine 11
	if result[0].Status != "error" || result[0].Code != "inventory_in_file_twice" || result[0].Count != 2 || result[0].FirstLine != 11 {
		t.Errorf("first row mismatch: got %+v", result[0])
	}

	// Second row: name_empty error with Count 1
	if result[1].Status != "error" || result[1].Code != "name_empty" || result[1].Count != 1 || result[1].FirstLine != 7 {
		t.Errorf("second row mismatch: got %+v", result[1])
	}

	// Third row: no_author warning with Count 3 and FirstLine 4
	if result[2].Status != "warning" || result[2].Code != "no_author" || result[2].Count != 3 || result[2].FirstLine != 4 {
		t.Errorf("third row mismatch: got %+v", result[2])
	}
}

func TestTrimDetailsKeepsErrorsFirst(t *testing.T) {
	
	
	makeReport := func() *model.ImportReport {
		return &model.ImportReport{
			Details: []model.ImportRowResult{
				{Status: "warning", Message: "warning 1", Line: 1},
				{Status: "warning", Message: "warning 2", Line: 2},
				{Status: "warning", Message: "warning 3", Line: 3},
				{Status: "warning", Message: "warning 4", Line: 4},
				{Status: "warning", Message: "warning 5", Line: 5},
				{Status: "warning", Message: "warning 6", Line: 6},
				{Status: "warning", Message: "warning 7", Line: 7},
				{Status: "warning", Message: "warning 8", Line: 8},
				{Status: "warning", Message: "warning 9", Line: 9},
				{Status: "warning", Message: "warning 10", Line: 10},
				{Status: "error", Message: "error 1", Line: 11},
				{Status: "error", Message: "error 2", Line: 12},
				{Status: "error", Message: "error 3", Line: 13},
			},
		}
	}

	
	rep := makeReport()
	TrimDetails(rep, 5)

	if rep.DetailsTotal != 13 {
		t.Errorf("expected DetailsTotal to be 13, got %d", rep.DetailsTotal)
	}

	if len(rep.Details) != 5 {
		t.Fatalf("expected 5 details, got %d", len(rep.Details))
	}

	// First 3 should be errors
	expectedLines := []int{11, 12, 13}
	for i, line := range expectedLines {
		if rep.Details[i].Line != line {
			t.Errorf("expected error line %d at position %d, got %d", line, i, rep.Details[i].Line)
		}
	}

	// Next 2 should be warnings
	expectedLines = []int{1, 2}
	for i, line := range expectedLines {
		if rep.Details[i+3].Line != line {
			t.Errorf("expected warning line %d at position %d, got %d", line, i+3, rep.Details[i+3].Line)
		}
	}

	
	repZero := makeReport()
	TrimDetails(repZero, 0)
	
	if repZero.DetailsTotal != 13 {
		t.Errorf("expected DetailsTotal to be 13 after limit 0, got %d", repZero.DetailsTotal)
	}
	if len(repZero.Details) != 13 {
		t.Errorf("expected 13 details after limit 0, got %d", len(repZero.Details))
	}
}
