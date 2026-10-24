package main

import (
	"fmt"
	"os"

	"libratrack/internal/repository"
	"libratrack/internal/service"
)

func main() {
	// 1. If os.Getenv("APP_ENV") == "production": fmt.Fprintln(os.Stderr, "refusing to seed in production"); os.Exit(1).
	if os.Getenv("APP_ENV") == "production" {
		fmt.Fprintln(os.Stderr, "refusing to seed in production")
		os.Exit(1)
	}

	// 2. dsn := os.Getenv("DATABASE_URL"). If dsn == "": print "DATABASE_URL is required" to stderr and exit 1. migrations := envOr("MIGRATIONS_PATH", "migrations").
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}
	migrations := envOr("MIGRATIONS_PATH", "migrations")

	// 3. Run the migrations and open the database exactly like cmd/api/main.go does. On error print it to stderr and exit 1.
	err := repository.RunMigrations(dsn, migrations)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	db, err := repository.NewPostgresDB(dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 4. password := os.Getenv("SEED_PASSWORD"). If password == "": generate one with randomPassword(16) and remember generated = true.
	password := os.Getenv("SEED_PASSWORD")
	generated := false
	if password == "" {
		password, err = randomPassword(16)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		generated = true
	}

	// 5. users := repository.NewUserRepository(db); loans := repository.NewLoanRepository(db).
	users := repository.NewUserRepository(db)
	loans := repository.NewLoanRepository(db)

	// 6. policy, err := service.NewLoanPolicy(envIntOr("LOAN_PERIOD_DAYS", 14), envIntOr("FINE_PER_DAY", 2), envIntOr("FINE_MAX", 100), envIntOr("MAX_ACTIVE_LOANS", 5)). On error exit 1.
	policy, err := service.NewLoanPolicy(
		envIntOr("LOAN_PERIOD_DAYS", 14),
		envIntOr("FINE_PER_DAY", 2),
		envIntOr("FINE_MAX", 100),
		envIntOr("MAX_ACTIVE_LOANS", 5),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 7. ids, reader1Existed, err := createUsers(users, password). On error exit 1.
	ids, reader1Existed, err := createUsers(users, password)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 8. If reader1Existed: fmt.Println("loans skipped: demo loans already exist"). Otherwise: createLoans(db, loans, policy, ids); on error exit 1.
	if !reader1Existed {
		err = createLoans(db, loans, policy, ids)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		fmt.Println("loans skipped: demo loans already exist")
	}

	// 9. If generated: print once to stdout "demo password: " + password (it is never written to a file or a log).
	if generated {
		fmt.Println("demo password: " + password)
	}

	// 10. fmt.Println("seed finished").
	fmt.Println("seed finished")
}
