package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/app/tenantlifecycle"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func runInspect(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	raw := flags.String("tenant", "", "full test tenant")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected inspection arguments")
	}
	key, err := tenant.ParseTestTenant(*raw)
	if err != nil {
		return err
	}
	cfg, err := postgresconfig.Load("TENANT_POSTGRES", "TENANT_POSTGRES_URL")
	if err != nil {
		return err
	}
	pool, err := postgres.Open(ctx, cfg.WithApplicationName("tadoku-tenant-inspect").URL().Reveal(), 1)
	if err != nil {
		return fmt.Errorf("open tenant postgres: %s", cfg.Redact(err))
	}
	defer pool.Close()
	exists, err := tenantlifecycle.NewApplication(pool, nil, nil, nil).Inspect(ctx, key)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Tenant string `json:"tenant"`
		Exists bool   `json:"exists"`
	}{key.String(), exists})
}
