package service

import "testing"

func TestNormalizeLanguage(t *testing.T) {
	// Table (input -> expected code, expected ok):
	//   "uk" -> "ukr", true;  "UKR" -> "ukr", true;  " ru " -> "rus", true;  "deu" -> "ger", true;  "fra" -> "fre", true;
	//   ""   -> "", true;     "xx" -> "", false;      "ukrainian" -> "", false.
	tests := []struct {
		input    string
		expected string
		ok       bool
	}{
		{"uk", "ukr", true},
		{"UKR", "ukr", true},
		{" ru ", "rus", true},
		{"deu", "ger", true},
		{"fra", "fre", true},
		{"", "", true},
		{"xx", "", false},
		{"ukrainian", "", false},
	}

	for _, tt := range tests {
		code, ok := NormalizeLanguage(tt.input)
		if code != tt.expected || ok != tt.ok {
			t.Errorf("NormalizeLanguage(%q) = (%q, %v), want (%q, %v)", tt.input, code, ok, tt.expected, tt.ok)
		}
	}
}

func TestLanguageTwoLetter(t *testing.T) {
	// "ukr" -> "uk"; "ger" -> "de"; "zzz" -> "".
	tests := []struct {
		input    string
		expected string
	}{
		{"ukr", "uk"},
		{"ger", "de"},
		{"zzz", ""},
	}

	for _, tt := range tests {
		code := LanguageTwoLetter(tt.input)
		if code != tt.expected {
			t.Errorf("LanguageTwoLetter(%q) = %q, want %q", tt.input, code, tt.expected)
		}
	}
}
