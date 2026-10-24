package service

import (
	"errors"
	"io"

	"github.com/xuri/excelize/v2"
)

// ReadXLSX reads the first sheet of an xlsx file (first row = header) into RawRows.
func ReadXLSX(r io.Reader) ([]RawRow, error) {
	f, err := excelize.OpenReader(r, excelize.Options{UnzipSizeLimit: 100 << 20, UnzipXMLSizeLimit: 50 << 20})
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("the workbook has no sheets")
	}
	rows, err := f.Rows(sheets[0])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	table := [][]string{}
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		table = append(table, cols)
		if len(table) > maxImportRows+1 {
			return nil, errors.New("too many rows")
		}
	}
	return rowsFromTable(table)
}
