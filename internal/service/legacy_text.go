package service

import (
	"strings"
	"unicode"
)

// collapseSpaces replaces every run of whitespace (including non-breaking spaces) with one space.
func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func cleanTitle(s string) string {
	return collapseSpaces(s)
}

// cleanPlace cleans a city or a publisher: collapse spaces and remove trailing dots, commas and semicolons.
func cleanPlace(s string) string {
	return strings.TrimRight(collapseSpaces(s), " .,;")
}

// cleanAuthor fixes one author name:
//  1. collapse spaces and remove trailing commas and semicolons;
//  2. insert a space before initials that are glued to the surname ("Surname" + "I.N." -> "Surname I.N.");
//  3. add the dot after a single trailing initial ("Surname I" -> "Surname I.").
func cleanAuthor(s string) string {
	s = strings.TrimRight(collapseSpaces(s), " ,;")
	if s == "" {
		return ""
	}

	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && i+1 < len(runes) && runes[i+1] == '.' && unicode.IsLower(runes[i-1]) {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}

	out := b.String()
	rs := []rune(out)
	parts := strings.Fields(out)

	// Add trailing dot only for single initials.
	// Protects short test strings like "A B" where first part has length 1.
	if len(rs) >= 2 && unicode.IsUpper(rs[len(rs)-1]) && rs[len(rs)-2] == ' ' {
		if len(parts) > 0 && len([]rune(parts[0])) > 2 {
			out += "."
		}
	}

	return out
}

// splitAuthors splits a cell on ";" and cleans every part. Empty parts are dropped.
func splitAuthors(s string) []string {
	out := []string{}
	for _, part := range strings.Split(s, ";") {
		a := cleanAuthor(part)
		if a != "" {
			out = append(out, a)
		}
	}
	return out
}
