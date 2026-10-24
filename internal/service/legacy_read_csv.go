package service

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
)

// ReadCSV reads a CSV file (first row = header) into RawRows.
func ReadCSV(r io.Reader) ([]RawRow, error) {
	// 1. data, err := io.ReadAll(r). Return the error if any.
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	// 2. data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}).
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	// 3. firstLine := string(data); if i := strings.IndexByte(firstLine, '\n'); i >= 0 { firstLine = firstLine[:i] }.
	firstLine := string(data)
	if i := strings.IndexByte(firstLine, '\n'); i >= 0 {
		firstLine = firstLine[:i]
	}
	// 4. comma := ','. If strings.Count(firstLine, ";") > strings.Count(firstLine, ",") { comma = ';' }.
	comma := ','
	if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
		comma = ';'
	}
	// 5. cr := csv.NewReader(bytes.NewReader(data)); cr.Comma = comma; cr.FieldsPerRecord = -1; cr.LazyQuotes = true.
	cr := csv.NewReader(bytes.NewReader(data))
	cr.Comma = comma
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	// 6. table, err := cr.ReadAll(). Return the error if any.
	table, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	// 7. Return rowsFromTable(table).
	return rowsFromTable(table)
}
