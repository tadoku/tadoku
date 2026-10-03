package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/app/tenantlifecycle"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/valkey"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type command struct {
	action    string
	key       tenant.TestKey
	testers   []uuid.UUID
	features  string
	component string
}

type testerIDs []uuid.UUID

func (testers *testerIDs) String() string { return "tester UUID" }

func (testers *testerIDs) Set(raw string) error {
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return errors.New("tester must be a nonzero UUID")
	}
	*testers = append(*testers, id)
	return nil
}

func parseCommand(args []string) (command, error) {
	if len(args) == 0 {
		return command{}, errors.New("usage: tadoku-tenant provision|teardown|override set|clear --tenant <name/route>")
	}
	action := args[0]
	args = args[1:]
	if action == "override" {
		if len(args) == 0 || (args[0] != "set" && args[0] != "clear") {
			return command{}, errors.New("override requires set or clear")
		}
		action += " " + args[0]
		args = args[1:]
	}
	if action != "provision" && action != "teardown" && action != "override set" && action != "override clear" {
		return command{}, errors.New("unknown tenant command")
	}
	flags := flag.NewFlagSet(action, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	rawTenant := flags.String("tenant", "", "full parsed test tenant")
	var result command
	result.action = action
	var testers testerIDs
	if action == "provision" {
		flags.StringVar(&result.features, "flipt-features", "", "Flipt seed file")
		flags.Var(&testers, "tester", "tester UUID; repeat to grant several identities")
	}
	if action == "override set" || action == "override clear" {
		flags.StringVar(&result.component, "component", "", "registered worker component")
	}
	if err := flags.Parse(args); err != nil {
		return command{}, err
	}
	if flags.NArg() != 0 {
		return command{}, errors.New("unexpected tenant command arguments")
	}
	key, err := tenant.ParseTestTenant(*rawTenant)
	if err != nil {
		return command{}, err
	}
	result.key = key
	result.testers = testers
	if action == "provision" && result.features == "" {
		return command{}, errors.New("provision requires --flipt-features")
	}
	if (action == "override set" || action == "override clear") && result.component != jobqueue.WorkerComponent {
		return command{}, errors.New("override component is not registered")
	}
	return result, nil
}

type config struct {
	KetoReadURL        string `envconfig:"keto_read_url" required:"true"`
	KetoWriteURL       string `envconfig:"keto_write_url" required:"true"`
	ValkeyURL          string `envconfig:"valkey_url" required:"true"`
	FliptManagementURL string `envconfig:"flipt_management_url" required:"true"`
	FliptEnvironment   string `envconfig:"flipt_environment" default:"test"`
}

func run(ctx context.Context, args []string) error {
	command, err := parseCommand(args)
	if err != nil {
		return err
	}
	var features []fliptmanagement.Resource
	if command.action == "provision" {
		file, err := os.Open(command.features)
		if err != nil {
			return fmt.Errorf("open Flipt features: %w", err)
		}
		features, err = fliptmanagement.ParseFeatures(file)
		closeErr := file.Close()
		if err := errors.Join(err, closeErr); err != nil {
			return fmt.Errorf("parse Flipt features: %w", err)
		}
	}
	postgresConfig, err := postgresconfig.Load("TENANT_POSTGRES", "TENANT_POSTGRES_URL")
	if err != nil {
		return err
	}
	pool, err := postgres.Open(ctx, postgresConfig.WithApplicationName("tadoku-tenant").URL().Reveal(), 1)
	if err != nil {
		return fmt.Errorf("open tenant postgres: %s", postgresConfig.Redact(err))
	}
	defer pool.Close()
	application := tenantlifecycle.NewApplication(pool, nil, nil, nil)
	if command.action == "override set" {
		return application.SetOverride(ctx, command.key, command.component)
	}
	if command.action == "override clear" {
		return application.ClearOverride(ctx, command.key, command.component)
	}

	cfg := config{}
	if err := envconfig.Process("TENANT", &cfg); err != nil {
		return err
	}
	for _, raw := range []string{cfg.KetoReadURL, cfg.KetoWriteURL, cfg.FliptManagementURL} {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return errors.New("tenant provider URL must be absolute HTTP(S) without credentials, query or fragment")
		}
	}
	targets, err := flipt.NewTargets("production", "default", cfg.FliptEnvironment)
	if err != nil {
		return err
	}
	client, err := valkey.Open(ctx, cfg.ValkeyURL, time.Second)
	if client != nil {
		defer client.Close()
	}
	if err != nil {
		return err
	}
	httpClient := &http.Client{Timeout: 30 * time.Second}
	defer httpClient.CloseIdleConnections()
	application = tenantlifecycle.NewApplication(
		pool,
		permissions.NewTenantManager(keto.NewClient(cfg.KetoReadURL, cfg.KetoWriteURL, keto.WithHTTPClient(httpClient))),
		leaderboard.NewCache(client, time.Second),
		fliptmanagement.NewClient(fliptmanagement.Config{
			URL: cfg.FliptManagementURL, Targets: targets, HTTPClient: httpClient,
		}),
	)
	if command.action == "provision" {
		return application.Provision(ctx, command.key, command.testers, features)
	}
	return application.Teardown(ctx, command.key)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
