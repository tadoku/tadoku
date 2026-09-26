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

func (a *Application) Run(ctx context.Context) error {
	poll := time.NewTicker(500 * time.Millisecond)
	defer poll.Stop()
	for {
		if ctx.Err() != nil {
			return nil
		}
		count, err := a.leaderboard.ReconcileCache(ctx)
		if err == nil {
			a.runner.logger.Info("leaderboard cache reconciled", "invalidated", count)
			break
		}
		a.runner.logger.Error("reconcile leaderboard cache", "error", err)
		select {
		case <-ctx.Done():
			return nil
		case <-poll.C:
		}
	}

	a.runner.run(ctx)
	return nil
}

func Replay(ctx context.Context, queue *jobqueue.Service, id int64, actor, reason string) (int64, error) {
	handlers, err := new(Application).registrations()
	if err != nil {
		return 0, err
	}
	return queue.Replay(ctx, id, actor, reason, handlers.types())
}
