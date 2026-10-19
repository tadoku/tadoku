package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz/tokens"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func runAccessToken(ctx context.Context, args []string) error {
	if len(args) == 0 || (args[0] != "create" && args[0] != "list" && args[0] != "revoke") {
		return errors.New("access-token requires create, list or revoke")
	}
	action := args[0]
	flags := flag.NewFlagSet("access-token "+action, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	rawTenant := flags.String("tenant", "", "full tenant key")
	var request tokens.Create
	var scope, rawID string
	if action == "create" {
		flags.StringVar(&request.Name, "name", "", "token name")
		flags.StringVar(&scope, "scope", "", "email:read")
		flags.DurationVar(&request.ExpiresIn, "expires-in", 0, "positive duration up to 720h")
	}
	if action == "revoke" {
		flags.StringVar(&rawID, "id", "", "token UUID")
	}
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected access-token arguments")
	}
	key, err := tenant.Parse(*rawTenant)
	if err != nil {
		return err
	}
	var id uuid.UUID
	if action == "revoke" {
		id, err = uuid.Parse(rawID)
		if err != nil || id == uuid.Nil {
			return errors.New("revoke requires a nonzero --id UUID")
		}
	}
	cfg := struct {
		Production bool `envconfig:"production" default:"true"`
	}{}
	if err := envconfig.Process("TENANT", &cfg); err != nil {
		return err
	}
	if action == "create" {
		if cfg.Production && key == tenant.Production() {
			return errors.New("access tokens are unavailable for tadoku/prod in production")
		}
		if request.ExpiresIn > 30*24*time.Hour {
			return errors.New("token expiry must not exceed 30 days")
		}
		request.Scopes = []tokens.Scope{tokens.Scope(scope)}
	}
	postgresConfig, err := postgresconfig.Load("TENANT_POSTGRES", "TENANT_POSTGRES_URL")
	if err != nil {
		return err
	}
	pool, err := postgres.Open(ctx, postgresConfig.WithApplicationName("tadoku-tenant-access-token").URL().Reveal(), 1)
	if err != nil {
		return fmt.Errorf("open access token postgres: %s", postgresConfig.Redact(err))
	}
	defer pool.Close()
	service := tokens.NewService(tokens.NewRepository(pool), cfg.Production)
	ctx = tenant.WithKey(ctx, key)
	switch action {
	case "create":
		var issued tokens.Issued
		err := postgres.RunInTransaction(ctx, pool, func(ctx context.Context) error {
			var err error
			issued, err = service.Create(ctx, request)
			return err
		})
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(os.Stdout, issued.Secret)
		return err
	case "list":
		rows, err := service.List(ctx)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(rows)
	default:
		return service.Revoke(ctx, id)
	}
}
