package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func run(ctx context.Context, args []string, output io.Writer) error {
	flags := flag.NewFlagSet("tadoku-identity-import", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	apply := flags.Bool("apply", false, "commit the import; otherwise only print differences")
	dryRun := flags.Bool("dry-run", false, "explicitly request the default dry run")
	expect := flags.Int("expect-count", 0, "expected current source identity count; required with --apply")
	if err := flags.Parse(args); errors.Is(err, flag.ErrHelp) {
		_, err := fmt.Fprintln(output,
			"usage: tadoku-identity-import [--dry-run | --apply --expect-count N]\n"+
				"Configure IMPORT_SOURCE_POSTGRES_* and IMPORT_TARGET_POSTGRES_*; target tenant is tadoku/prod.")
		return err
	} else if err != nil || flags.NArg() != 0 {
		return errors.New("invalid import arguments; use --help")
	}
	if *expect < 0 || (*apply && (*dryRun || *expect == 0)) {
		return errors.New("apply requires a positive --expect-count and cannot use --dry-run")
	}

	sourceConfig, err := postgresconfig.Load("IMPORT_SOURCE_POSTGRES", "IMPORT_SOURCE_POSTGRES_URL")
	if err != nil {
		return err
	}
	targetConfig, err := postgresconfig.Load("IMPORT_TARGET_POSTGRES", "IMPORT_TARGET_POSTGRES_URL")
	if err != nil {
		return err
	}
	if sourceConfig.Host == targetConfig.Host && sourceConfig.Port == targetConfig.Port &&
		sourceConfig.Database == targetConfig.Database {
		return errors.New("source and target must be distinct databases")
	}

	source, err := postgres.Open(
		ctx,
		sourceConfig.WithApplicationName("tadoku-identity-import-source").URL().Reveal(),
		1,
	)
	if err != nil {
		return databaseError("source connection", err)
	}
	defer source.Close()

	target, err := postgres.Open(
		ctx,
		targetConfig.WithApplicationName("tadoku-identity-import-target").URL().Reveal(),
		1,
	)
	if err != nil {
		return databaseError("target connection", err)
	}
	defer target.Close()
	return identity.ImportKratosAccounts(ctx, source, target, *expect, *apply, output)
}

func databaseError(stage string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return fmt.Errorf("%s failed (SQLSTATE %s); database details suppressed", stage, pgError.Code)
	}
	return fmt.Errorf("%s failed; connection and database details suppressed", stage)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
