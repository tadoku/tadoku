package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	smtpinfra "github.com/tadoku/tadoku/services/tadoku-api/infra/smtp"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestEmailWorkerReconciliation(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			api.reset(t, "")
			fixture := emailSMTPServer(t, false)
			var sender email.Sender
			if enabled {
				var err error
				sender, err = smtpinfra.New(fixture.config)
				if err != nil {
					t.Fatal(err)
				}
			}
			service := email.NewPersistentService(email.NewRepository(api.db.AppPool), sender)
			ctx := tenant.WithKey(t.Context(), tenant.Production())
			ids := make(map[string]uuid.UUID)
			for _, name := range []string{"failed", "abandoned", "pending replay", "running replay"} {
				id := writeEmail(t, ctx, service, recoveryMail())
				ids[name] = id
				var jobID int64
				err := api.db.Pool.QueryRow(ctx, `update jobs
					set state = 'failed', attempts = 5, failed_at = now(), last_error = 'lease_expired'
					where tenant = 'tadoku/prod' and payload->>'email_id' = $1
					returning id`, id.String()).Scan(&jobID)
				if err != nil {
					t.Fatal(err)
				}

				if name == "abandoned" {
					_, err = api.db.Pool.Exec(ctx, `update jobs
						set state = 'running', failed_at = null, last_error = null,
							claim_token = $2, lease_expires_at = now() - interval '1 minute'
						where id = $1`, jobID, uuid.New())
				} else if name == "pending replay" || name == "running replay" {
					_, err = api.db.Pool.Exec(ctx, `insert into jobs (
						tenant, task_type, payload, replay_of_id, replay_actor, replay_reason, next_attempt_at
					)
					select tenant, task_type, payload, id, 'fixture', 'repaired', now() + interval '1 day'
					from jobs where id = $1`, jobID)
					if err == nil && name == "running replay" {
						_, err = api.db.Pool.Exec(ctx, `update jobs
							set state = 'running', claim_token = $2, lease_expires_at = now() + interval '1 day'
							where replay_of_id = $1`, jobID, uuid.New())
					}
				}
				if err != nil {
					t.Fatal(err)
				}
			}

			scope, _ := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
			application := emailWorker(t, service, jobqueue.NewRepository(api.db.AppPool), scope)
			workerCtx, cancel := context.WithCancel(t.Context())
			done := make(chan error, 1)
			go func() { done <- application.Run(workerCtx) }()
			defer func() {
				cancel()
				if err := <-done; err != nil {
					t.Error(err)
				}
			}()

			var want email.Delivery = email.Held{}
			if enabled {
				want = email.Failed{Error: "lease_expired"}
			}
			deadline := time.Now().Add(20 * time.Second)
			for {
				row, err := service.Find(ctx, ids["abandoned"])
				if err != nil {
					t.Fatal(err)
				}
				if row.Delivery == want {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("abandoned final lease: delivery=%+v want %+v", row.Delivery, want)
				}
				time.Sleep(20 * time.Millisecond)
			}

			for name, id := range ids {
				row, err := service.Find(ctx, id)
				if err != nil {
					t.Fatal(err)
				}
				expected := want
				if strings.Contains(name, "replay") {
					expected = email.Pending{}
				}
				if row.Delivery != expected {
					t.Errorf("%s delivery=%+v want %+v", name, row.Delivery, expected)
				}
				if failed, ok := row.Delivery.(email.Failed); ok && failed.Error != "lease_expired" {
					t.Errorf("%s failure=%q want lease_expired", name, failed.Error)
				}
			}
			fixture.mu.Lock()
			defer fixture.mu.Unlock()
			if fixture.attempts != 0 {
				t.Errorf("maintenance made %d SMTP attempts", fixture.attempts)
			}
		})
	}
}

func emailWorker(
	t *testing.T,
	service *email.Service,
	repository *jobqueue.Repository,
	scope jobqueue.Scope,
) *worker.Application {
	t.Helper()
	application, err := worker.NewApplication(jobqueue.NewService(repository), api.leaderboard, worker.Config{
		Email:           service,
		Scope:           scope,
		Concurrency:     1,
		ShutdownTimeout: time.Second,
		Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return application
}

type emailClaimPause struct {
	once    sync.Once
	started chan struct{}
	resume  chan struct{}
}

func (p *emailClaimPause) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return ctx
}

func (*emailClaimPause) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (p *emailClaimPause) TraceBatchStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	for _, query := range data.Batch.QueuedQueries {
		if strings.HasPrefix(query.SQL, "-- name: Claim :") {
			p.once.Do(func() {
				close(p.started)
				<-p.resume
			})
		}
	}
	return ctx
}

func (*emailClaimPause) TraceBatchQuery(context.Context, *pgx.Conn, pgx.TraceBatchQueryData) {}

func (*emailClaimPause) TraceBatchEnd(context.Context, *pgx.Conn, pgx.TraceBatchEndData) {}

func TestEmailWorkerStartupAndDailyRetention(t *testing.T) {
	api.reset(t, "")
	synctest.Test(t, func(t *testing.T) {
		ctx := tenant.WithKey(t.Context(), tenant.Production())
		service := emailService()
		startupID := writeEmail(t, ctx, service, recoveryMail())
		dailyID := writeEmail(t, ctx, service, recoveryMail())
		for id, age := range map[uuid.UUID]time.Duration{
			startupID: 4 * 24 * time.Hour,
			dailyID:   2*24*time.Hour + time.Hour,
		} {
			_, err := api.db.Pool.Exec(ctx, `update emails set created_at = $2, status = 'held' where id = $1`,
				id, timex.Now().Add(-age))
			if err != nil {
				t.Fatal(err)
			}
		}

		pause := &emailClaimPause{started: make(chan struct{}), resume: make(chan struct{})}
		config := api.db.AppPool.Config()
		config.ConnConfig.Tracer = pause
		pool, err := pgxpool.NewWithConfig(t.Context(), config)
		if err != nil {
			t.Fatal(err)
		}
		defer pool.Close()
		scope, _ := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
		application := emailWorker(t, service, jobqueue.NewRepository(pool), scope)
		workerCtx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() { done <- application.Run(workerCtx) }()
		defer func() {
			cancel()
			if err := <-done; err != nil {
				t.Error(err)
			}
		}()

		<-pause.started
		var startupExists, dailyExists bool
		read := func() {
			t.Helper()
			err := api.db.Pool.QueryRow(ctx, `select
				exists(select 1 from emails where id = $1), exists(select 1 from emails where id = $2)`,
				startupID, dailyID).Scan(&startupExists, &dailyExists)
			if err != nil {
				t.Fatal(err)
			}
		}
		read()
		if startupExists || !dailyExists {
			t.Errorf("startup retention: expired exists=%t, recent exists=%t", startupExists, dailyExists)
		}

		time.Sleep(24 * time.Hour)
		close(pause.resume)
		synctest.Wait()
		read()
		if dailyExists {
			t.Error("daily retention left expired email after the timer fired")
		}
	})
}
