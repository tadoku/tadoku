package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/app/accountdeletion"
	"github.com/tadoku/tadoku/services/tadoku-api/app/permissionmodel"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contestevents"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/observability"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	valkeyinfra "github.com/tadoku/tadoku/services/tadoku-api/infra/valkey"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type config struct {
	Email                  emailConfig
	Branch                 string `envconfig:"branch"`
	Component              string `envconfig:"component"`
	scope                  jobqueue.Scope
	Concurrency            int                   `validate:"gt=0" envconfig:"concurrency" default:"4"`
	Port                   int                   `validate:"gt=0,lte=65535" default:"8000"`
	MetricsPort            int                   `validate:"gt=0,lte=65535" envconfig:"metrics_port" default:"9090"`
	Postgres               postgresconfig.Config `ignored:"true"`
	PostgresMaxConnections int32                 `validate:"gt=0,lte=32" envconfig:"postgres_max_connections" default:"4"`
	ValkeyURL              string                `validate:"required" envconfig:"valkey_url"`
	ValkeyTimeout          time.Duration         `validate:"gt=0" envconfig:"valkey_timeout" default:"1s"`
	DialTimeout            time.Duration         `validate:"gt=0" envconfig:"dial_timeout" default:"3s"`
	ShutdownTimeout        time.Duration         `validate:"gt=0" envconfig:"shutdown_timeout" default:"15s"`
}

func loadConfig() (config, error) {
	cfg := config{Component: jobqueue.WorkerComponent}
	if err := envconfig.Process("WORKER", &cfg); err != nil {
		return config{}, fmt.Errorf("load worker config: %w", err)
	}

	deployment, err := tenant.ParseDeployment(cfg.Branch)
	if err != nil {
		return config{}, fmt.Errorf("validate worker config: WORKER_BRANCH: %w", err)
	}
	cfg.scope, err = jobqueue.AllTenantsExcept(cfg.Component)
	if err != nil {
		return config{}, fmt.Errorf("validate worker config: WORKER_COMPONENT: %w", err)
	}
	if key, branch := deployment.Key(); branch {
		cfg.scope, err = jobqueue.OnlyTenant(key)
		if err != nil {
			return config{}, fmt.Errorf("validate worker config: WORKER_BRANCH: %w", err)
		}
	}

	if err := validator.New().Struct(cfg); err != nil {
		return config{}, fmt.Errorf("validate worker config: %w", err)
	}

	cfg.Postgres, err = postgresconfig.Load("WORKER_POSTGRES", "WORKER_POSTGRES_URL")
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}

func run(ctx context.Context, cfg config, logger *slog.Logger) error {
	sender, err := cfg.Email.sender()
	if err != nil {
		return fmt.Errorf("configure worker email: %w", err)
	}
	startupCtx, cancelStartup := context.WithTimeout(ctx, cfg.DialTimeout)
	pool, err := postgres.Open(
		startupCtx,
		cfg.Postgres.WithApplicationName("tadoku-worker").URL().Reveal(),
		cfg.PostgresMaxConnections,
	)
	cancelStartup()
	if err != nil {
		return fmt.Errorf("open worker postgres: %s", cfg.Postgres.Redact(err))
	}
	defer pool.Close()

	client, valkeyErr := valkeyinfra.Open(ctx, cfg.ValkeyURL, cfg.ValkeyTimeout)
	if client == nil {
		return valkeyErr
	}
	defer client.Close()
	if valkeyErr != nil {
		logger.Warn("valkey unavailable at startup; retrying in worker", "error", valkeyErr)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		postgres.NewPoolCollector(pool),
	)

	metrics := worker.NewMetrics(registry)
	model, err := permissionmodel.New()
	if err != nil {
		return fmt.Errorf("build permission model: %w", err)
	}
	application, err := worker.NewApplication(
		worker.Dependencies{
			JobQueue: jobqueue.NewService(jobqueue.NewRepository(pool)),
			Leaderboard: leaderboard.NewService(
				leaderboard.NewRepository(pool),
				leaderboard.NewCache(client, cfg.ValkeyTimeout),
			),
			Email:          email.NewService(email.NewRepository(pool), sender),
			IdentityExpiry: identity.NewExpiry(identity.NewRepository(pool)),
		},
		worker.Config{
			AccountDeletion: &worker.AccountDeletion{
				DB:            pool,
				Profile:       profile.NewService(profile.NewRepository(pool), identity.NewReader(pool)),
				Logs:          logs.NewService(logs.NewLogsRepository(pool), false),
				Contests:      contests.NewService(contests.NewContestsRepository(pool)),
				Relationships: permissions.NewRelationships(authz.NewRepository(pool), model),
				Identity:      identity.NewDeletion(identity.NewRepository(pool)),
				Audit:         audit.NewService(audit.NewRepository(pool)),
				ContestEvents: contestevents.NewService(contestevents.NewRepository(pool)),
			},
			Scope:           cfg.scope,
			Concurrency:     cfg.Concurrency,
			ShutdownTimeout: cfg.ShutdownTimeout,
			Logger:          logger,
			Metrics:         metrics,
		},
	)
	if err != nil {
		return fmt.Errorf("construct worker application: %w", err)
	}

	privateMux := http.NewServeMux()
	privateMux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	privateMux.HandleFunc("GET /readyz", func(w http.ResponseWriter, request *http.Request) {
		if !application.Ready() || pool.Ping(request.Context()) != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	privateServer := &http.Server{Handler: privateMux, ReadHeaderTimeout: 5 * time.Second}
	metricsServer := &http.Server{Handler: promhttp.HandlerFor(registry, promhttp.HandlerOpts{}), ReadHeaderTimeout: 5 * time.Second}

	privateListener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		return fmt.Errorf("listen for worker health: %w", err)
	}
	defer privateListener.Close()
	metricsListener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.MetricsPort))
	if err != nil {
		return fmt.Errorf("listen for worker metrics: %w", err)
	}
	defer metricsListener.Close()

	workCtx, cancelWork := context.WithCancel(ctx)
	workerDone := make(chan error, 1)
	go func() { workerDone <- application.Run(workCtx) }()

	serverErrors := make(chan error, 2)
	go func() {
		if err := privateServer.Serve(privateListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("worker health server: %w", err)
		}
	}()
	go func() {
		if err := metricsServer.Serve(metricsListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("worker metrics server: %w", err)
		}
	}()

	logger.Info("tadoku-worker started", "address", privateListener.Addr(), "postgres_max_connections", cfg.PostgresMaxConnections)

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-serverErrors:
	}

	cancelWork()
	runErr = errors.Join(runErr, <-workerDone)

	shutdownCtx, stop := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer stop()
	return errors.Join(runErr, privateServer.Shutdown(shutdownCtx), metricsServer.Shutdown(shutdownCtx))
}

func replay(ctx context.Context, args []string, logger *slog.Logger) error {
	ctx = tenant.WithKey(ctx, tenant.Production())

	flags := flag.NewFlagSet("replay", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	id := flags.Int64("id", 0, "failed job ID")
	actor := flags.String("actor", "", "operator identity")
	reason := flags.String("reason", "", "reason for replay")

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse replay flags: %w", err)
	}
	if flags.NArg() != 0 || *id < 1 || strings.TrimSpace(*actor) == "" || strings.TrimSpace(*reason) == "" {
		return errors.New("usage: tadoku-worker replay --id <failed-id> --actor <actor> --reason <reason>")
	}

	postgresConfig, err := postgresconfig.Load("WORKER_POSTGRES", "WORKER_POSTGRES_URL")
	if err != nil {
		return err
	}

	openCtx, stop := context.WithTimeout(ctx, 3*time.Second)
	pool, err := postgres.Open(openCtx, postgresConfig.WithApplicationName("tadoku-worker-replay").URL().Reveal(), 1)
	stop()
	if err != nil {
		return fmt.Errorf("open replay postgres: %s", postgresConfig.Redact(err))
	}
	defer pool.Close()

	newID, err := worker.Replay(ctx, jobqueue.NewService(jobqueue.NewRepository(pool)), *id, *actor, *reason)
	if err != nil {
		return fmt.Errorf("replay job %d: %w", *id, err)
	}
	logger.InfoContext(ctx, "job replayed", "source_job_id", *id, "job_id", newID)
	return nil
}

func deleteAccount(ctx context.Context, args []string, logger *slog.Logger) error {
	ctx = tenant.WithKey(ctx, tenant.Production())

	flags := flag.NewFlagSet("delete-account", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	user := flags.String("user", "", "user ID to delete")
	actor := flags.String("actor", "", "operator user ID")
	reason := flags.String("reason", "", "reason for the deletion")

	const usage = "usage: tadoku-worker delete-account --user <user-id> --actor <operator-user-id> --reason <reason>"
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse delete-account flags: %w", err)
	}
	userID, userErr := uuid.Parse(*user)
	actorID, actorErr := uuid.Parse(*actor)
	if flags.NArg() != 0 || userErr != nil || actorErr != nil {
		return errors.New(usage)
	}
	request, err := accountdeletion.OperatorRequest(userID, actorID, *reason)
	if err != nil {
		return fmt.Errorf("%s: %w", usage, err)
	}

	postgresConfig, err := postgresconfig.Load("WORKER_POSTGRES", "WORKER_POSTGRES_URL")
	if err != nil {
		return err
	}
	openCtx, stop := context.WithTimeout(ctx, 3*time.Second)
	pool, err := postgres.Open(
		openCtx,
		postgresConfig.WithApplicationName("tadoku-worker-delete-account").URL().Reveal(),
		1,
	)
	stop()
	if err != nil {
		return fmt.Errorf("open delete-account postgres: %s", postgresConfig.Redact(err))
	}
	defer pool.Close()

	model, err := permissionmodel.New()
	if err != nil {
		return err
	}
	tuples := authz.NewRepository(pool)
	filing := accountdeletion.NewFiling(
		authz.NewRoles(permissions.NewChecker(tuples, model), permissions.NewRelationships(tuples, model)),
		identity.NewDeletion(identity.NewRepository(pool)),
		profile.NewService(profile.NewRepository(pool), identity.NewReader(pool)),
		audit.NewService(audit.NewRepository(pool)),
		jobqueue.NewService(jobqueue.NewRepository(pool)),
	)
	jobID, err := worker.FileAccountDeletion(ctx, pool, filing, request)
	if err != nil {
		return fmt.Errorf("delete account %s: %w", userID, err)
	}
	logger.InfoContext(ctx, "account deletion requested", "user_id", userID, "job_id", jobID)
	return nil
}

func main() {
	logger := slog.New(observability.NewTenantHandler(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if len(os.Args) > 1 {
		var err error
		switch os.Args[1] {
		case "replay":
			err = replay(ctx, os.Args[2:], logger)
		case "delete-account":
			err = deleteAccount(ctx, os.Args[2:], logger)
		default:
			err = errors.New("usage: tadoku-worker [replay --id <failed-id> --actor <actor> --reason <reason>" +
				" | delete-account --user <user-id> --actor <operator-user-id> --reason <reason>]")
		}
		if err != nil {
			logger.Error("tadoku-worker command failed", "error", err)
			os.Exit(1)
		}
		return
	}
	cfg, err := loadConfig()
	if err == nil {
		err = run(ctx, cfg, logger)
	}
	if err != nil {
		logger.Error("tadoku-worker failed", "error", err)
		os.Exit(1)
	}
}
