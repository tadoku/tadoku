package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type migration interface {
	Version() (version uint, dirty bool, err error)
	Force(version int) error
	Close() (sourceErr error, databaseErr error)
}

type migrationFactory func(sourceURL, databaseURL string) (migration, error)

func newMigration(sourceURL, databaseURL string) (migration, error) {
	return migrate.New(sourceURL, databaseURL)
}

func closeMigration(runner migration, stderr io.Writer, redact func(any) string) bool {
	sourceErr, databaseErr := runner.Close()
	if sourceErr != nil {
		fmt.Fprintf(stderr, "migrate-recovery: close source: %s\n", redact(sourceErr))
	}
	if databaseErr != nil {
		fmt.Fprintf(stderr, "migrate-recovery: close database: %s\n", redact(databaseErr))
	}
	return sourceErr == nil && databaseErr == nil
}

func run(args []string, stdout, stderr io.Writer, factory migrationFactory) int {
	flags := flag.NewFlagSet("migrate-recovery", flag.ContinueOnError)
	flags.SetOutput(stderr)

	sourceURL := flags.String("source", "", "migration source URL")
	expectedVersion := flags.Int("expected-version", 0, "dirty version observed during inspection")
	targetVersion := flags.Int("target-version", 0, "version matching the verified physical schema")
	confirmation := flags.String("confirm-target-version", "", "repeat target version to confirm metadata repair")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *sourceURL == "" {
		fmt.Fprintln(stderr, "migrate-recovery: -source is required")
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(stderr, "migrate-recovery: expected command: inspect or force")
		return 2
	}

	command := flags.Arg(0)
	if command != "inspect" && command != "force" {
		fmt.Fprintln(stderr, "migrate-recovery: expected command: inspect or force")
		return 2
	}

	setFlags := map[string]bool{}
	flags.Visit(func(current *flag.Flag) {
		setFlags[current.Name] = true
	})

	if command == "inspect" &&
		(setFlags["expected-version"] || setFlags["target-version"] || setFlags["confirm-target-version"]) {
		fmt.Fprintln(stderr, "migrate-recovery: force flags are not valid with inspect")
		return 2
	}
	if command == "force" {
		for _, name := range []string{"expected-version", "target-version", "confirm-target-version"} {
			if !setFlags[name] {
				fmt.Fprintf(stderr, "migrate-recovery: -%s is required with force\n", name)
				return 2
			}
		}
		if *targetVersion < -1 {
			fmt.Fprintln(stderr, "migrate-recovery: -target-version must be at least -1")
			return 2
		}
		if *expectedVersion < 0 {
			fmt.Fprintln(stderr, "migrate-recovery: -expected-version must be non-negative")
			return 2
		}
		if *targetVersion > *expectedVersion {
			fmt.Fprintln(stderr, "migrate-recovery: -target-version cannot exceed -expected-version")
			return 2
		}
		if *confirmation != strconv.Itoa(*targetVersion) {
			fmt.Fprintln(stderr, "migrate-recovery: confirmation does not match target version")
			return 2
		}
	}

	postgresConfig, err := postgresconfig.Load("POSTGRES", "POSTGRES_URL")
	if err != nil {
		fmt.Fprintf(stderr, "migrate-recovery: postgres configuration: %v\n", err)
		return 2
	}
	databaseURL := postgresConfig.URL()
	redact := func(value any) string {
		return postgresConfig.Redact(value)
	}
	runner, err := factory(*sourceURL, databaseURL)
	if err != nil {
		fmt.Fprintf(stderr, "migrate-recovery: initialize: %s\n", redact(err))
		return 1
	}

	version, dirty, versionErr := runner.Version()
	if versionErr != nil && !errors.Is(versionErr, migrate.ErrNilVersion) {
		fmt.Fprintf(stderr, "migrate-recovery: inspect: %s\n", redact(versionErr))
		closeMigration(runner, stderr, redact)
		return 1
	}

	if command == "inspect" {
		if errors.Is(versionErr, migrate.ErrNilVersion) {
			fmt.Fprintln(stdout, "version=nil dirty=false")
		} else {
			fmt.Fprintf(stdout, "version=%d dirty=%t\n", version, dirty)
		}
		if !closeMigration(runner, stderr, redact) {
			return 1
		}
		return 0
	}

	if errors.Is(versionErr, migrate.ErrNilVersion) {
		fmt.Fprintln(stderr, "migrate-recovery: refusing force: database has no migration version")
		closeMigration(runner, stderr, redact)
		return 1
	}
	if !dirty {
		fmt.Fprintf(stderr, "migrate-recovery: refusing force: version %d is not dirty\n", version)
		closeMigration(runner, stderr, redact)
		return 1
	}
	if int(version) != *expectedVersion {
		fmt.Fprintf(
			stderr,
			"migrate-recovery: refusing force: expected dirty version %d, found %d\n",
			*expectedVersion,
			version,
		)
		closeMigration(runner, stderr, redact)
		return 1
	}

	if err := runner.Force(*targetVersion); err != nil {
		fmt.Fprintf(stderr, "migrate-recovery: force: %s\n", redact(err))
		closeMigration(runner, stderr, redact)
		return 1
	}

	forcedVersion, forcedDirty, forcedVersionErr := runner.Version()
	if *targetVersion == -1 {
		if !errors.Is(forcedVersionErr, migrate.ErrNilVersion) || forcedDirty {
			fmt.Fprintln(stderr, "migrate-recovery: force verification failed")
			closeMigration(runner, stderr, redact)
			return 1
		}
	} else if forcedVersionErr != nil || forcedDirty || int(forcedVersion) != *targetVersion {
		fmt.Fprintln(stderr, "migrate-recovery: force verification failed")
		closeMigration(runner, stderr, redact)
		return 1
	}

	if !closeMigration(runner, stderr, redact) {
		return 1
	}
	fmt.Fprintf(stdout, "forced version=%d dirty=false\n", *targetVersion)
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, newMigration))
}
