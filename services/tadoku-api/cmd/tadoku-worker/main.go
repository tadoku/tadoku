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
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	valkeyinfra "github.com/tadoku/tadoku/services/tadoku-api/infra/valkey"
)

type config struct {
	Port                   int                   `validate:"gt=0,lte=65535" default:"8000"`
	MetricsPort            int                   `validate:"gt=0,lte=65535" envconfig:"metrics_port" default:"9090"`
	Postgres               postgresconfig.Config `ignored:"true"`
	PostgresMaxConnections int32                 `validate:"gt=0,lte=32" envconfig:"postgres_max_connections" default:"4"`
	ValkeyURL              string                `validate:"required" envconfig:"valkey_url"`
	ValkeyTimeout          time.Duration         `validate:"gt=0" envconfig:"valkey_timeout" default:"1s"`
	LeaderboardCachePrefix string                `envconfig:"leaderboard_cache_prefix"`
	DialTimeout            time.Duration         `validate:"gt=0" envconfig:"dial_timeout" default:"3s"`
	ShutdownTimeout        time.Duration         `validate:"gt=0" envconfig:"shutdown_timeout" default:"15s"`
}

func loadConfig() (config, error) {
	cfg := config{}
	if err := envconfig.Process("WORKER", &cfg); err != nil {
		return config{}, fmt.Errorf("load worker config: %w", err)
	}
	if err := validator.New().Struct(cfg); err != nil {
		return config{}, fmt.Errorf("validate worker config: %w", err)
	}
	if prefix := cfg.LeaderboardCachePrefix; prefix != "" {
		if !strings.HasSuffix(prefix, ":") || strings.IndexFunc(prefix, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == ':')
		}) >= 0 {
			return config{}, errors.New("WORKER_LEADERBOARD_CACHE_PREFIX must contain only lowercase letters, digits, hyphens and colons, and end in a colon")
		}
	}
	var err error
	cfg.Postgres, err = postgresconfig.Load("WORKER_POSTGRES", "WORKER_POSTGRES_URL")
	if err != nil {
		return config{}, err
	}
	return cfg, nil
}

func run(ctx context.Context, cfg config, logger *slog.Logger) error {
	startupCtx, cancelStartup := context.WithTimeout(ctx, cfg.DialTimeout)
	pool, err := postgres.Open(startupCtx, cfg.Postgres.WithApplicationName("tadoku-worker").URL(), cfg.PostgresMaxConnections)
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
	leaderboardService := leaderboard.NewService(leaderboard.NewRepository(pool), client, cfg.ValkeyTimeout, cfg.LeaderboardCachePrefix)
	application, err := worker.NewApplication(jobqueue.NewService(jobqueue.NewRepository(pool)), leaderboardService, worker.Config{Logger: logger, Metrics: metrics, ShutdownTimeout: cfg.ShutdownTimeout})
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
	flags := flag.NewFlagSet("replay", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	id := flags.Int64("id", 0, "failed task ID")
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
	pool, err := postgres.Open(openCtx, postgresConfig.WithApplicationName("tadoku-worker-replay").URL(), 1)
	stop()
	if err != nil {
		return fmt.Errorf("open replay postgres: %s", postgresConfig.Redact(err))
	}
	defer pool.Close()
	newID, err := worker.Replay(ctx, jobqueue.NewService(jobqueue.NewRepository(pool)), *id, *actor, *reason)
	if err != nil {
		return fmt.Errorf("replay task %d: %w", *id, err)
	}
	logger.Info("async task replayed", "source_id", *id, "task_id", newID)
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if len(os.Args) > 1 {
		var err error
		if os.Args[1] == "replay" {
			err = replay(ctx, os.Args[2:], logger)
		} else {
			err = errors.New("usage: tadoku-worker [replay --id <failed-id> --actor <actor> --reason <reason>]")
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
