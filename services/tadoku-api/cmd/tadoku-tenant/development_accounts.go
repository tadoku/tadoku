package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/app/tenantlifecycle"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func runDevelopmentAccounts(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("development-accounts", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	rawTenant := flags.String("tenant", "", "full test tenant")
	admin := flags.String("admin-id", "", "existing development administrator UUID")
	reader := flags.String("reader-id", "", "existing development reader UUID")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected development account arguments")
	}
	key, err := tenant.ParseTestTenant(*rawTenant)
	if err != nil {
		return err
	}
	adminID, err := uuid.Parse(*admin)
	if err != nil || adminID == uuid.Nil {
		return errors.New("development administrator UUID is required")
	}
	readerID, err := uuid.Parse(*reader)
	if err != nil || readerID == uuid.Nil || readerID == adminID {
		return errors.New("distinct development reader UUID is required")
	}
	cfg := struct {
		Production bool `envconfig:"production" default:"true"`
	}{}
	if err := envconfig.Process("TENANT", &cfg); err != nil {
		return err
	}
	if cfg.Production {
		return errors.New("development accounts are unavailable in production")
	}

	postgresConfig, err := postgresconfig.Load("TENANT_POSTGRES", "TENANT_POSTGRES_URL")
	if err != nil {
		return err
	}
	host := os.Getenv("TENANT_POSTGRES_HOST")
	if host != "tadoku-dev-db.tdk-dev-data" && host != "127.0.0.1" && host != "localhost" {
		return errors.New("development accounts require the development database service or loopback")
	}
	pool, err := postgres.Open(ctx, postgresConfig.WithApplicationName("tadoku-development-accounts").URL().Reveal(), 1)
	if err != nil {
		return fmt.Errorf("open development postgres: %s", postgresConfig.Redact(err))
	}
	defer pool.Close()

	defaults := func(key, value string) string {
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		return value
	}
	fixtures := []identity.DevelopmentAccount{
		{ID: adminID, Email: defaults("TADOKU_DEV_ADMIN_EMAIL", "dev@tadoku.app"), DisplayName: "Dev Admin",
			Password: defaults("TADOKU_DEV_ADMIN_PASSWORD", "tadoku")},
		{ID: readerID, Email: defaults("TADOKU_DEV_READER_EMAIL", "reader@tadoku.app"), DisplayName: "Dev Reader",
			Password: defaults("TADOKU_DEV_READER_PASSWORD", "tadoku")},
	}
	return tenantlifecycle.NewApplication(pool, nil, nil, nil).SeedDevelopmentAccounts(ctx, key, fixtures)
}
