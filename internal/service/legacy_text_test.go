package service

import "testing"

func TestCleanAuthor(t *testing.T) {
	// Table (input -> expected):
	//   "\u0410\u0432\u0435\u0440\u044c\u044f\u043d\u043e\u0432\u041e.\u0418."  -> "\u0410\u0432\u0435\u0440\u044c\u044f\u043d\u043e\u0432 \u041e.\u0418."   (space inserted before the initials)
	//   "\u0410\u0431\u0434\u0443\u043b\u0430\u0435\u0432 \u0418,"              -> "\u0410\u0431\u0434\u0443\u043b\u0430\u0435\u0432 \u0418."             (trailing comma removed, dot added)
	//   "\u0410\u0431\u0440\u0430\u043c\u043e\u0432 \u0422.\u0413."              -> "\u0410\u0431\u0440\u0430\u043c\u043e\u0432 \u0422.\u0413."             (unchanged)
	//   "   " -> ""
	tests := []struct {
		input    string
		expected string
	}{
		{"\u0410\u0432\u0435\u0440\u044c\u044f\u043d\u043e\u0432\u041e.\u0418.", "\u0410\u0432\u0435\u0440\u044c\u044f\u043d\u043e\u0432 \u041e.\u0418."},
		{"\u0410\u0431\u0434\u0443\u043b\u0430\u0435\u0432 \u0418,", "\u0410\u0431\u0434\u0443\u043b\u0430\u0435\u0432 \u0418."},
		{"\u0410\u0431\u0440\u0430\u043c\u043e\u0432 \u0422.\u0413.", "\u0410\u0431\u0440\u0430\u043c\u043e\u0432 \u0422.\u0413."},
		{"   ", ""},
	}

	for _, tt := range tests {
		result := cleanAuthor(tt.input)
		if result != tt.expected {
			t.Errorf("cleanAuthor(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestCleanPlace(t *testing.T) {
	// "\u0420\u043e\u0441\u0442\u043e\u0432 \u043d\u0430 \u0414\u043e\u043d\u0443." -> "\u0420\u043e\u0441\u0442\u043e\u0432 \u043d\u0430 \u0414\u043e\u043d\u0443"
	// "\u0412\u044b\u0441\u0448\u0430\u044f \u0448\u043a\u043e\u043b\u0430." -> "\u0412\u044b\u0441\u0448\u0430\u044f \u0448\u043a\u043e\u043b\u0430"
	// "  X  Y ." -> "X Y"
	tests := []struct {
		input    string
		expected string
	}{
		{"\u0420\u043e\u0441\u0442\u043e\u0432 \u043d\u0430 \u0414\u043e\u043d\u0443.", "\u0420\u043e\u0441\u0442\u043e\u0432 \u043d\u0430 \u0414\u043e\u043d\u0443"},
		{"\u0412\u044b\u0441\u0448\u0430\u044f \u0448\u043a\u043e\u043b\u0430.", "\u0412\u044b\u0441\u0448\u0430\u044f \u0448\u043a\u043e\u043b\u0430"},
		{"  X  Y .", "X Y"},
	}

	for _, tt := range tests {
		result := cleanPlace(tt.input)
		if result != tt.expected {
			t.Errorf("cleanPlace(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSplitAuthors(t *testing.T) {
	// "A B; C D," gives two names. "" gives an empty slice.
	tests := []struct {
		input    string
		expected []string
	}{
		{"A B; C D,", []string{"A B", "C D"}},
		{"", []string{}},
	}

	for _, tt := range tests {
		result := splitAuthors(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitAuthors(%q) = %v, want %v", tt.input, result, tt.expected)
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("splitAuthors(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}
