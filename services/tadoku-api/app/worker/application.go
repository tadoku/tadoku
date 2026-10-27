package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contestevents"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type Dependencies struct {
	DB               *pgxpool.Pool
	JobQueue         *jobqueue.Service
	Leaderboard      *leaderboard.Service
	Email            *email.Service
	IdentityExpiry   *identity.Expiry
	IdentityDeletion *identity.Deletion
	Profile          *profile.Service
	Logs             *logs.Service
	Contests         *contests.Service
	Relationships    *permissions.Relationships
	Audit            *audit.Service
	ContestEvents    *contestevents.Service
}

type Config struct {
	Scope           jobqueue.Scope
	Concurrency     int
	ShutdownTimeout time.Duration
	Logger          *slog.Logger
	Metrics         *Metrics
}

type Application struct {
	runner           *runner
	db               *pgxpool.Pool
	queue            *jobqueue.Service
	leaderboard      *leaderboard.Service
	email            *email.Service
	identityDeletion *identity.Deletion
	profile          *profile.Service
	logs             *logs.Service
	contests         *contests.Service
	relationships    *permissions.Relationships
	audit            *audit.Service
	contestEvents    *contestevents.Service
}

func NewApplication(deps Dependencies, config Config) (*Application, error) {
	if err := requireAll(deps); err != nil {
		return nil, err
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

	a := &Application{
		db:               deps.DB,
		queue:            deps.JobQueue,
		leaderboard:      deps.Leaderboard,
		email:            deps.Email,
		identityDeletion: deps.IdentityDeletion,
		profile:          deps.Profile,
		logs:             deps.Logs,
		contests:         deps.Contests,
		relationships:    deps.Relationships,
		audit:            deps.Audit,
		contestEvents:    deps.ContestEvents,
	}
	handlers, err := registrations(
		a.InvalidateContestLeaderboard,
		a.InvalidateOfficialLeaderboard,
		a.SendEmail,
		a.DeleteAccount,
	)
	if err != nil {
		return nil, err
	}

	a.runner = &runner{
		scope:           config.Scope,
		email:           deps.Email,
		identity:        deps.IdentityExpiry,
		queue:           deps.JobQueue,
		handlers:        handlers,
		logger:          config.Logger,
		metrics:         config.Metrics,
		shutdownTimeout: config.ShutdownTimeout,
		concurrency:     config.Concurrency,
	}
	a.runner.metrics.initialize(handlers)
	return a, nil
}

func requireAll(deps Dependencies) error {
	switch {
	case deps.DB == nil:
		return errors.New("worker requires DB")
	case deps.JobQueue == nil:
		return errors.New("worker requires JobQueue")
	case deps.Leaderboard == nil:
		return errors.New("worker requires Leaderboard")
	case deps.Email == nil:
		return errors.New("worker requires Email")
	case deps.IdentityExpiry == nil:
		return errors.New("worker requires IdentityExpiry")
	case deps.IdentityDeletion == nil:
		return errors.New("worker requires IdentityDeletion")
	case deps.Profile == nil:
		return errors.New("worker requires Profile")
	case deps.Logs == nil:
		return errors.New("worker requires Logs")
	case deps.Contests == nil:
		return errors.New("worker requires Contests")
	case deps.Relationships == nil:
		return errors.New("worker requires Relationships")
	case deps.Audit == nil:
		return errors.New("worker requires Audit")
	case deps.ContestEvents == nil:
		return errors.New("worker requires ContestEvents")
	}
	return nil
}

func registrations(
	invalidateContest func(context.Context, jobs.InvalidateContestLeaderboardV1) error,
	invalidateOfficial func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error,
	sendEmail func(context.Context, jobs.SendEmailV1) error,
	deleteAccount func(context.Context, jobs.DeleteAccountV1) error,
) (*registry, error) {
	leaderboardPolicy := policy{
		Concurrency: 2,
		Timeout:     20 * time.Second,
		MaxAttempts: 5,
	}
	return newRegistry(
		handle(invalidateContest, leaderboardPolicy),
		handle(invalidateOfficial, leaderboardPolicy),
		handle(sendEmail, policy{Concurrency: 2, Timeout: 20 * time.Second, MaxAttempts: 5}),
		handle(deleteAccount, policy{Concurrency: 1, Timeout: 60 * time.Second, MaxAttempts: 5}),
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
	handlers, err := registrations(
		ignoreJob[jobs.InvalidateContestLeaderboardV1],
		ignoreJob[jobs.InvalidateOfficialLeaderboardV1],
		ignoreJob[jobs.SendEmailV1],
		ignoreJob[jobs.DeleteAccountV1],
	)
	if err != nil {
		return 0, err
	}
	return queue.Replay(ctx, id, actor, reason, handlers.types())
}

func ignoreJob[J jobs.Job](context.Context, J) error { return nil }

func (a *Application) SendEmail(ctx context.Context, job jobs.SendEmailV1) error {
	final, _ := ctx.Value(finalAttemptKey{}).(bool)
	return a.email.SendAttempt(ctx, job.EmailID, final)
}
