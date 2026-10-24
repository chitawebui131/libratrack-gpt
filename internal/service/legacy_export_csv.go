package service

import (
	"encoding/csv"
	"io"

	"libratrack/internal/model"
)

func csvSafe(cell string) string {
	// If cell != "" and its first byte is one of '=', '+', '-', '@', '\t', '\r', return "'" + cell. Otherwise return cell.
	if cell != "" {
		switch cell[0] {
		case '=', '+', '-', '@', '\t', '\r':
			return "'" + cell
		}
	}
	return cell
}

// WriteCSV writes the books in the same 14 columns as the import file.
func WriteCSV(w io.Writer, rows []model.ExportRow) error {
	// 1. Write the BOM bytes 0xEF, 0xBB, 0xBF with w.Write. Return the error if any.
	_, err := w.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		return err
	}
	// 2. cw := csv.NewWriter(w); cw.Comma = ';'.
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	// 3. Write the header legacyHeaders with cw.Write. Return the error if any.
	err = cw.Write(legacyHeaders)
	if err != nil {
		return err
	}
	// 4. For every row with len(row.Inventory) > 0: cells := exportCells(row); for i := range cells { cells[i] = csvSafe(cells[i]) }; cw.Write(cells); return the error if any.
	for _, row := range rows {
		if len(row.Inventory) > 0 {
			cells := exportCells(row)
			for i := range cells {
				cells[i] = csvSafe(cells[i])
			}
			err = cw.Write(cells)
			if err != nil {
				return err
			}
		}
	}
	// 5. cw.Flush(). Return cw.Error().
	cw.Flush()
	return cw.Error()
}
