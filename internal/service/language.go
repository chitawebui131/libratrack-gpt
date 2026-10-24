package service

import (
	_ "embed"
	"encoding/csv"
	"strings"
	"sync"
)

//go:embed languages.csv
var languagesCSV string

var (
	langOnce  sync.Once
	langTable map[string]string // any accepted spelling -> ISO 639-2/B code
	langTwo   map[string]string // ISO 639-2/B code -> ISO 639-1 code
)

func loadLanguages() {
	// 1. langTable = map[string]string{}; langTwo = map[string]string{}.
	langTable = map[string]string{}
	langTwo = map[string]string{}
	// 2. records, err := csv.NewReader(strings.NewReader(languagesCSV)).ReadAll(). If err != nil, return.
	records, err := csv.NewReader(strings.NewReader(languagesCSV)).ReadAll()
	if err != nil {
		return
	}
	// 3. For every record at index >= 1 with len(record) >= 3: two := lower-case trimmed record[0]; three := lower-case trimmed record[1].
	//      langTable[two] = three; langTable[three] = three; langTwo[three] = two.
	//      For every alias in strings.Split(record[2], ";"): alias = lower-case trimmed; if alias != "" { langTable[alias] = three }.
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) >= 3 {
			two := strings.ToLower(strings.TrimSpace(record[0]))
			three := strings.ToLower(strings.TrimSpace(record[1]))
			langTable[two] = three
			langTable[three] = three
			langTwo[three] = two
			for _, alias := range strings.Split(record[2], ";") {
				alias = strings.ToLower(strings.TrimSpace(alias))
				if alias != "" {
					langTable[alias] = three
				}
			}
		}
	}
}

// NormalizeLanguage returns the ISO 639-2/B code (3 letters) for a 2-letter code, a 3-letter code or a known alias.
// An empty input returns "" and ok = true. An unknown value returns "" and ok = false.
func NormalizeLanguage(s string) (string, bool) {
	// 1. langOnce.Do(loadLanguages). s = strings.ToLower(strings.TrimSpace(s)). If s == "", return "", true.
	langOnce.Do(loadLanguages)
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", true
	}
	// 2. code, ok := langTable[s]. Return code, ok.
	code, ok := langTable[s]
	return code, ok
}

// LanguageTwoLetter returns the ISO 639-1 code for an ISO 639-2/B code, or "" when there is none (used later for Dublin Core).
func LanguageTwoLetter(code3 string) string {
	// langOnce.Do(loadLanguages); return langTwo[strings.ToLower(code3)].
	langOnce.Do(loadLanguages)
	return langTwo[strings.ToLower(code3)]
}
