package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
)

type Config struct {
	Concurrency     int
	ShutdownTimeout time.Duration
	Logger          *slog.Logger
	Metrics         *Metrics
}

type Application struct {
	runner      *runner
	leaderboard *leaderboard.Service
}

func NewApplication(queue *jobqueue.Service, leaderboard *leaderboard.Service, config Config) (*Application, error) {
	if queue == nil || leaderboard == nil {
		return nil, errors.New("worker requires queue and leaderboard services")
	}
	if config.Concurrency == 0 {
		config.Concurrency = 4
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 15 * time.Second
	}
	if config.Concurrency < 1 || config.ShutdownTimeout <= 0 {
		return nil, errors.New("worker concurrency and shutdown timeout must be positive")
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.Metrics == nil {
		config.Metrics = NewMetrics(prometheus.NewRegistry())
	}
	a := &Application{leaderboard: leaderboard}
	handlers, err := a.registrations()
	if err != nil {
		return nil, err
	}
	a.runner = &runner{
		queue:           queue,
		handlers:        handlers,
		logger:          config.Logger,
		metrics:         config.Metrics,
		shutdownTimeout: config.ShutdownTimeout,
		concurrency:     config.Concurrency,
	}
	a.runner.metrics.initialize(handlers)
	return a, nil
}

func (a *Application) registrations() (*registry, error) {
	return newRegistry(
		handle(a.InvalidateContestLeaderboard, Policy{Concurrency: 2, Timeout: 20 * time.Second, MaxAttempts: 5}),
		handle(a.InvalidateOfficialLeaderboard, Policy{Concurrency: 2, Timeout: 20 * time.Second, MaxAttempts: 5}),
	)
}

func (a *Application) InvalidateContestLeaderboard(ctx context.Context, job jobs.InvalidateContestLeaderboardV1) error {
	return a.leaderboard.InvalidateContest(ctx, job.ContestID)
}

func (a *Application) InvalidateOfficialLeaderboard(ctx context.Context, job jobs.InvalidateOfficialLeaderboardV1) error {
	return a.leaderboard.InvalidateOfficial(ctx, job.Year)
}

func (a *Application) Ready() bool { return a.runner.ready.Load() }

// Run must be called once per application. Handlers must respect context cancellation;
// an occupied execution slot is retained until its handler returns.
func (a *Application) Run(ctx context.Context) error {
	if err := a.leaderboard.RevokeCacheReadiness(ctx); err != nil {
		a.runner.logger.Warn("revoke leaderboard readiness at startup", "error", err)
	}
	defer func() {
		shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if err := a.leaderboard.RevokeCacheReadiness(shutdownCtx); err != nil {
			a.runner.logger.Warn("revoke leaderboard readiness at shutdown", "error", err)
		}
	}()
	reconciled := false
	a.runner.run(ctx, func(ctx context.Context, idle, failed bool) {
		if failed {
			reconciled = false
		}
		if idle && !reconciled {
			count, err := a.leaderboard.ReconcileCache(ctx)
			if err != nil {
				a.runner.logger.Error("reconcile leaderboard cache", "error", err)
			} else {
				reconciled = true
				a.runner.logger.Info("leaderboard cache reconciled", "invalidated", count)
			}
		}
		if idle && reconciled {
			if err := a.refreshReadiness(ctx); err != nil {
				a.runner.logger.Error("refresh leaderboard readiness", "error", err)
			}
		} else if err := a.leaderboard.RevokeCacheReadiness(ctx); err != nil {
			a.runner.logger.Warn("revoke leaderboard readiness while work is active", "error", err)
		}
	})
	return nil
}

func (a *Application) refreshReadiness(ctx context.Context) error {
	unsupported, err := a.runner.queue.UnsupportedStats(ctx, a.runner.handlers.types())
	if err != nil {
		_ = a.leaderboard.RevokeCacheReadiness(ctx)
		return err
	}
	if unsupported.Pending+unsupported.Running+unsupported.Failed > 0 {
		return a.leaderboard.RevokeCacheReadiness(ctx)
	}
	for _, name := range a.runner.handlers.types() {
		count, err := a.runner.queue.Outstanding(ctx, name)
		if err != nil {
			_ = a.leaderboard.RevokeCacheReadiness(ctx)
			return err
		}
		if count > 0 {
			return a.leaderboard.RevokeCacheReadiness(ctx)
		}
	}
	return a.leaderboard.PublishCacheReadiness(ctx)
}

// Replay uses this executable's registered versions and requires only PostgreSQL.
func Replay(ctx context.Context, queue *jobqueue.Service, id int64, actor, reason string) (int64, error) {
	handlers, err := new(Application).registrations()
	if err != nil {
		return 0, err
	}
	return queue.Replay(ctx, id, actor, reason, handlers.types())
}
