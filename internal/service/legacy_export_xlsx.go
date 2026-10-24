package service

import (
	"io"
	"strconv"

	"github.com/xuri/excelize/v2"

	"libratrack/internal/model"
)

// WriteXLSX writes the books in the same 14 columns as the import file, on the sheet "Sheet1".
func WriteXLSX(w io.Writer, rows []model.ExportRow) error {
	// 1. f := excelize.NewFile(); defer f.Close(). const sheet = "Sheet1".
	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Sheet1"

	// 2. header := make([]interface{}, len(legacyHeaders)) filled from legacyHeaders. f.SetSheetRow(sheet, "A1", &header). Return the error if any.
	header := make([]interface{}, len(legacyHeaders))
	for i, h := range legacyHeaders {
		header[i] = h
	}
	err := f.SetSheetRow(sheet, "A1", &header)
	if err != nil {
		return err
	}

	// 3. line := 2. For every row with len(row.Inventory) > 0:
	//      cells := exportCells(row); values := make([]interface{}, len(cells)).
	//      For i, c := range cells: values[i] = c. Then convert the numeric columns: indexes 0, 8 and 10 (code, countPage, countBooks): if cells[i] != "" { n, err := strconv.Atoi(cells[i]); if err == nil { values[i] = n } }.
	//      cell, err := excelize.CoordinatesToCellName(1, line). Return the error if any. f.SetSheetRow(sheet, cell, &values). Return the error if any. line++.
	line := 2
	for _, row := range rows {
		if len(row.Inventory) <= 0 {
			continue
		}
		cells := exportCells(row)
		values := make([]interface{}, len(cells))
		for i, c := range cells {
			values[i] = c
		}
		// Convert numeric columns: indexes 0, 8 and 10 (code, countPage, countBooks)
		for _, i := range []int{0, 8, 10} {
			if cells[i] != "" {
				n, err := strconv.Atoi(cells[i])
				if err == nil {
					values[i] = n
				}
			}
		}
		cell, err := excelize.CoordinatesToCellName(1, line)
		if err != nil {
			return err
		}
		err = f.SetSheetRow(sheet, cell, &values)
		if err != nil {
			return err
		}
		line++
	}

	// 4. Return f.Write(w).
	return f.Write(w)
}
