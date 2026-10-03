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
	Scope           jobqueue.Scope
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
	if config.Concurrency < 1 || config.ShutdownTimeout <= 0 {
		return nil, errors.New("worker requires positive concurrency and shutdown timeout")
	}

	if _, err := config.Scope.Context(context.Background()); err != nil {
		return nil, err
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.Metrics == nil {
		config.Metrics = NewMetrics(prometheus.NewRegistry())
	}

	a := &Application{leaderboard: leaderboard}
	handlers, err := registrations(a.InvalidateContestLeaderboard, a.InvalidateOfficialLeaderboard)
	if err != nil {
		return nil, err
	}

	a.runner = &runner{
		scope:           config.Scope,
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

func registrations(
	invalidateContest func(context.Context, jobs.InvalidateContestLeaderboardV1) error,
	invalidateOfficial func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error,
) (*registry, error) {
	leaderboardPolicy := policy{
		Concurrency: 2,
		Timeout:     20 * time.Second,
		MaxAttempts: 5,
	}
	return newRegistry(
		handle(invalidateContest, leaderboardPolicy),
		handle(invalidateOfficial, leaderboardPolicy),
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
	a.runner.run(ctx)
	return nil
}

func Replay(ctx context.Context, queue *jobqueue.Service, id int64, actor, reason string) (int64, error) {
	handlers, err := registrations(ignoreJob[jobs.InvalidateContestLeaderboardV1], ignoreJob[jobs.InvalidateOfficialLeaderboardV1])
	if err != nil {
		return 0, err
	}
	return queue.Replay(ctx, id, actor, reason, handlers.types())
}

func ignoreJob[J jobs.Job](context.Context, J) error { return nil }
