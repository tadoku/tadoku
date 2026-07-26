package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type migration interface {
	Up() error
	Close() (sourceErr error, databaseErr error)
}

type migrationFactory func(sourceURL, databaseURL string) (migration, error)

func newMigration(sourceURL, databaseURL string) (migration, error) {
	return migrate.New(sourceURL, databaseURL)
}

func run(args []string, stdout, stderr io.Writer, factory migrationFactory) int {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	flags.SetOutput(stderr)

	sourceURL := flags.String("source", "", "migration source URL")
	databaseURL := flags.String("database", "", "database URL")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *sourceURL == "" {
		fmt.Fprintln(stderr, "migrate: -source is required")
		return 2
	}
	if *databaseURL == "" {
		fmt.Fprintln(stderr, "migrate: -database is required")
		return 2
	}
	if flags.NArg() != 1 || flags.Arg(0) != "up" {
		fmt.Fprintln(stderr, "migrate: expected command: up")
		return 2
	}

	runner, err := factory(*sourceURL, *databaseURL)
	if err != nil {
		fmt.Fprintf(stderr, "migrate: initialize: %v\n", err)
		return 1
	}

	migrationErr := runner.Up()
	sourceCloseErr, databaseCloseErr := runner.Close()

	if migrationErr != nil && !errors.Is(migrationErr, migrate.ErrNoChange) {
		fmt.Fprintf(stderr, "migrate: up: %v\n", migrationErr)
		return 1
	}
	if sourceCloseErr != nil {
		fmt.Fprintf(stderr, "migrate: close source: %v\n", sourceCloseErr)
		return 1
	}
	if databaseCloseErr != nil {
		fmt.Fprintf(stderr, "migrate: close database: %v\n", databaseCloseErr)
		return 1
	}

	if errors.Is(migrationErr, migrate.ErrNoChange) {
		fmt.Fprintln(stdout, "no change")
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, newMigration))
}
