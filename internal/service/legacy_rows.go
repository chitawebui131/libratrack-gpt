package service

import (
	"errors"
	"unicode/utf8"

	"libratrack/internal/model"
)

func tooLong(s string, max int) bool {
	// Return utf8.RuneCountInString(s) > max.
	return utf8.RuneCountInString(s) > max
}

// rowToImportBook converts the canonical cells of one row.
func rowToImportBook(cells map[string]string) (*model.ImportBook, []string, error) {
	// 1. warnings := []string{}.
	warnings := []string{}
	// 2. title := cleanTitle(cells[colName]). If title == "", return nil, nil, errors.New("name is empty"). If tooLong(title, 255), return an error "name is longer than 255 characters". If utf8.RuneCountInString(title) == 50, append the warning "title may be cut in the source (50 characters)".
	title := cleanTitle(cells[colName])
	if title == "" {
		return nil, nil, errors.New("name is empty")
	}
	if tooLong(title, 255) {
		return nil, nil, errors.New("name is longer than 255 characters")
	}
	if utf8.RuneCountInString(title) == 50 {
		warnings = append(warnings, "title may be cut in the source (50 characters)")
	}
	// 3. count, err := parseCount(cells[colCountBooks]). Return the error if any.
	count, err := parseCount(cells[colCountBooks])
	if err != nil {
		return nil, nil, err
	}
	// 4. inventory, invWarnings, err := parseInventar(cells[colInventar], count). Return the error if any. Append invWarnings to warnings.
	inventory, invWarnings, err := parseInventar(cells[colInventar], count)
	if err != nil {
		return nil, nil, err
	}
	warnings = append(warnings, invWarnings...)
	// 5. authors := splitAuthors(cells[colAuthor]). If len(authors) == 0: authors = []string{UnknownAuthorName()} and append the warning "no author".
	authors := splitAuthors(cells[colAuthor])
	if len(authors) == 0 {
		authors = []string{UnknownAuthorName()}
		warnings = append(warnings, "no author")
	}
	// 6. year, w := parseYear(cells[colYear]). If w != "", append it.
	year, w := parseYear(cells[colYear])
	if w != "" {
		warnings = append(warnings, w)
	}
	// 7. city := cleanPlace(cells[colCity]); publisher := cleanPlace(cells[colPublishing]); department := collapseSpaces(cells[colDepartment]); bbk := collapseSpaces(cells[colBBK]).
	city := cleanPlace(cells[colCity])
	publisher := cleanPlace(cells[colPublishing])
	department := collapseSpaces(cells[colDepartment])
	bbk := collapseSpaces(cells[colBBK])
	// 8. Length checks (return an error "<field> is too long"): city > 100, publisher > 255, department > 20, bbk > 30.
	if len(city) > 100 {
		return nil, nil, errors.New("city is too long")
	}
	if len(publisher) > 255 {
		return nil, nil, errors.New("publisher is too long")
	}
	if len(department) > 20 {
		return nil, nil, errors.New("department is too long")
	}
	if tooLong(bbk, 30) {
		return nil, nil, errors.New("bbk is too long")
	}
	// 9. code := parseInt0(cells[colCode]). udk := collapseSpaces(cells[colUDK]); if tooLong(udk, 100), return an error "udk is too long".
	code := parseInt0(cells[colCode])
	udk := collapseSpaces(cells[colUDK])
	if tooLong(udk, 100) {
		return nil, nil, errors.New("udk is too long")
	}
	// 9a. isbn, err := NormalizeISBN(cells[colISBN]). If err != nil: isbn = "" and append the warning "ISBN ignored: invalid".
	isbn, err := NormalizeISBN(cells[colISBN])
	if err != nil {
		isbn = ""
		warnings = append(warnings, "ISBN ignored: invalid")
	}
	// 9b. lang, ok := NormalizeLanguage(cells[colLanguage]). If !ok: lang = "" and append the warning "language ignored: unknown".
	lang, ok := NormalizeLanguage(cells[colLanguage])
	if !ok {
		lang = ""
		warnings = append(warnings, "language ignored: unknown")
	}
	// 10. (department is applied to the copies of this row by the repository.) Return &model.ImportBook{LegacyCode: code, Title: title, Authors: authors, Year: year, City: city, Publisher: publisher, Department: department, Pages: parseInt0(cells[colCountPage]), BBK: bbk, UDK: udk, ISBN: isbn, Language: lang, CategoryName: CategoryFromBBK(bbk), Inventory: inventory}, warnings, nil.
	return &model.ImportBook{
		LegacyCode:   code,
		Title:        title,
		Authors:      authors,
		Year:         year,
		City:         city,
		Publisher:    publisher,
		Department:   department,
		Pages:        parseInt0(cells[colCountPage]),
		BBK:          bbk,
		UDK:          udk,
		ISBN:         isbn,
		Language:     lang,
		CategoryName: CategoryFromBBK(bbk),
		Inventory:    inventory,
	}, warnings, nil
}
