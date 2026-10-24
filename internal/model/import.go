package model

// ImportBook is one catalog row after cleaning.
type ImportBook struct {
	LegacyCode   int // 0 means "no code"
	Title        string
	Authors      []string
	Year         int
	City         string
	Publisher    string
	Department   string
	Pages        int
	BBK          string
	UDK          string
	ISBN         string // normalized, may be empty
	Language     string // ISO 639-2/B code, may be empty
	CategoryName string
	Inventory    []string
}

// ImportResult is what the repository did for one book.
type ImportResult struct {
	BookCreated   bool
	CopiesAdded   int
	CopiesSkipped int
}

// ImportRowResult is one line of the import report. Status: created, updated, warning, error.
type ImportRowResult struct {
	Line    int
	Status  string
	Message string
}

type ImportReport struct {
	DryRun        bool
	Rows          int
	BooksCreated  int
	BooksUpdated  int
	CopiesAdded   int
	CopiesSkipped int
	Errors        int
	Warnings      int
	Details       []ImportRowResult
	Summary       []ImportSummaryRow
	DetailsTotal  int // number of detail lines before TrimDetails cut the list
}

// ExportRow is one book with its authors and the inventory numbers of ONE department, ready for export.
type ExportRow struct {
	LegacyCode *int
	Title      string
	Authors    string
	Year       int
	City       string
	Publisher  string
	Department string
	Pages      int
	BBK        string
	UDK        string
	ISBN       string
	Language   string
	Inventory  []string
	Count      int
}

// ImportSummaryRow groups the report lines of one kind (status + code).
type ImportSummaryRow struct {
	Status    string // error or warning
	Code      string // short ASCII code of the reason
	Count     int
	FirstLine int    // first file line with this reason
	Example   string // message of that first line
}
