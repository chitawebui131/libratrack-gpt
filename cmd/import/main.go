package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"libratrack/internal/repository"
	"libratrack/internal/service"
)

func main() {
	// 1. file := flag.String("file", "", "path to .xlsx or .csv"); dry := flag.Bool("dry-run", false, "parse and validate only"); flag.Parse(). If *file == "": print "usage: import -file books.xlsx [-dry-run]" to stderr and exit 2.
	file := flag.String("file", "", "path to .xlsx or .csv")
	dry := flag.Bool("dry-run", false, "parse and validate only")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, "usage: import -file books.xlsx [-dry-run]")
		os.Exit(2)
	}

	// 2. f, err := os.Open(*file). On error print it to stderr and exit 1. defer f.Close().
	f, err := os.Open(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	// 3. ext := strings.ToLower(filepath.Ext(*file)). Read with service.ReadXLSX(f) for ".xlsx" or service.ReadCSV(f) for ".csv"; any other extension: print an error and exit 2. On read error print it and exit 1.
	ext := strings.ToLower(filepath.Ext(*file))
	var rows []service.RawRow
	if ext == ".xlsx" {
		rows, err = service.ReadXLSX(f)
	} else if ext == ".csv" {
		rows, err = service.ReadCSV(f)
	} else {
		fmt.Fprintln(os.Stderr, "unsupported file extension")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 4. dsn := os.Getenv("DATABASE_URL"). If empty: print "DATABASE_URL is required" and exit 1. Open the database exactly like cmd/seed/main.go does (do NOT run the migrations here unless cmd/seed does it in the same place; the app applies them at startup).
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}
	db, err := repository.NewPostgresDB(dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 5. svc := service.NewImportService(repository.NewImportRepository(db)). report := svc.Run(rows, *dry).
	svc := service.NewImportService(repository.NewImportRepository(db))
	report := svc.Run(rows, *dry)

	// 6. Print a summary with fmt.Printf: rows, books created, books updated, copies added, copies skipped, errors, warnings. Then print every detail whose Status is "error" as "line N: message".
	fmt.Printf("rows: %d, books created: %d, books updated: %d, copies added: %d, copies skipped: %d, errors: %d, warnings: %d\n",
		report.Rows, report.BooksCreated, report.BooksUpdated, report.CopiesAdded, report.CopiesSkipped, report.Errors, report.Warnings)
	for _, detail := range report.Details {
		if detail.Status == "error" {
			fmt.Printf("line %d: %s\n", detail.Line, detail.Message)
		}
	}

	// 7. If report.Errors > 0, exit with code 3.
	if report.Errors > 0 {
		os.Exit(3)
	}
}
