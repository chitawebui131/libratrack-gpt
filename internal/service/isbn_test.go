package service

import (
	"errors"
	"testing"
)

func TestNormalizeISBNValid(t *testing.T) {
	// Table (input -> expected):
	//   "978-0-306-40615-7"       -> "9780306406157"
	//   "ISBN 978-0-306-40615-7"  -> "9780306406157"
	//   "0-306-40615-2"           -> "0306406152"      (ISBN-10 stays ISBN-10)
	//   "0-8044-2957-x"           -> "080442957X"      (lower-case x)
	//   "  "                      -> ""                (empty is allowed)
	tests := []struct {
		input    string
		expected string
	}{
		{"978-0-306-40615-7", "9780306406157"},
		{"ISBN 978-0-306-40615-7", "9780306406157"},
		{"0-306-40615-2", "0306406152"},
		{"0-8044-2957-x", "080442957X"},
		{"  ", ""},
	}

	for _, tt := range tests {
		result, err := NormalizeISBN(tt.input)
		if err != nil {
			t.Errorf("NormalizeISBN(%q) returned error: %v", tt.input, err)
		}
		if result != tt.expected {
			t.Errorf("NormalizeISBN(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeISBNInvalid(t *testing.T) {
	// Each of these returns an error that satisfies errors.Is(err, ErrInvalidISBN):
	//   "9780306406158" (wrong check digit), "0306406153" (wrong check digit), "12345", "97803064061X7", "abcdefghij".
	tests := []string{
		"9780306406158",
		"0306406153",
		"12345",
		"97803064061X7",
		"abcdefghij",
	}

	for _, input := range tests {
		_, err := NormalizeISBN(input)
		if err == nil {
			t.Errorf("NormalizeISBN(%q) should have returned an error", input)
		}
		if !errors.Is(err, ErrInvalidISBN) {
			t.Errorf("NormalizeISBN(%q) error %v does not satisfy errors.Is(err, ErrInvalidISBN)", input, err)
		}
	}
}
