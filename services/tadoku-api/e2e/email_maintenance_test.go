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
			service := email.NewService(email.NewRepository(api.db.AppPool), sender)
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

func TestEmailWorkerReconciliationRules(t *testing.T) {
	api.reset(t, "")
	fixture := emailSMTPServer(t, false)
	sender, err := smtpinfra.New(fixture.config)
	if err != nil {
		t.Fatal(err)
	}
	service := email.NewService(email.NewRepository(api.db.AppPool), sender)

	branch, _ := tenant.Parse("e2e/email-reconcile")
	_, err = api.db.Pool.Exec(t.Context(), "insert into tenants (key, kind) values ($1, 'test')", branch.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = api.db.Pool.Exec(context.Background(), "delete from tenants where key = $1", branch.String())
	})

	const failSend = `update jobs
		set state = 'failed', attempts = 5, failed_at = now(), last_error = 'handler_error'
		where tenant = $1 and task_type = 'email.send.v1' and payload->>'email_id' = $2`
	const failReplay = `insert into jobs (
			tenant, task_type, payload, state, attempts, failed_at, last_error,
			replay_of_id, replay_actor, replay_reason
		)
		select tenant, task_type, payload, 'failed', 5, now(), 'rejected', id, 'fixture', 'retried'
		from jobs where tenant = $1 and payload->>'email_id' = $2`
	const pendingOtherType = `insert into jobs (tenant, task_type, payload, next_attempt_at)
		select tenant, 'email.other.v1', payload, now() + interval '1 day'
		from jobs where tenant = $1 and payload->>'email_id' = $2`
	const extraFields = `update jobs
		set state = 'failed', attempts = 5, failed_at = now(), last_error = 'handler_error',
			payload = payload || '{"extra":true}'
		where tenant = $1 and payload->>'email_id' = $2`
	const failExtraReplay = `insert into jobs (
			tenant, task_type, payload, state, attempts, failed_at, last_error,
			replay_of_id, replay_actor, replay_reason
		)
		select tenant, task_type, payload || '{"extra":true}', 'failed', 5, now(), 'rejected', id, 'fixture', 'retried'
		from jobs where tenant = $1 and payload->>'email_id' = $2`
	const pendingExtraFields = `insert into jobs (tenant, task_type, payload, next_attempt_at)
		select tenant, task_type, payload || '{"extra":true}', now() + interval '1 day'
		from jobs where tenant = $1 and payload->>'email_id' = $2`
	const completedSend = `insert into jobs (tenant, task_type, payload, state, attempts, completed_at)
		select tenant, task_type, payload, 'completed', 1, now()
		from jobs where tenant = $1 and payload->>'email_id' = $2`
	cases := []struct {
		name     string
		key      tenant.Key
		fixtures []string
		want     email.Delivery
	}{
		{
			name: "missing error code",
			key:  tenant.Production(),
			fixtures: []string{`update jobs
				set state = 'failed', attempts = 5, failed_at = now(), last_error = null
				where tenant = $1 and payload->>'email_id' = $2`},
			want: email.Failed{Error: "attempts_exhausted"},
		},
		{
			name:     "latest failed replay",
			key:      tenant.Production(),
			fixtures: []string{failSend, failReplay},
			want:     email.Failed{Error: "rejected"},
		},
		{
			name:     "pending job of another type",
			key:      tenant.Production(),
			fixtures: []string{pendingOtherType, failSend},
			want:     email.Failed{Error: "handler_error"},
		},
		{
			name: "failed job of another type",
			key:  tenant.Production(),
			fixtures: []string{`update jobs
				set task_type = 'email.other.v1', state = 'failed', attempts = 5, failed_at = now(),
					last_error = 'handler_error'
				where tenant = $1 and payload->>'email_id' = $2`},
			want: email.Pending{},
		},
		{
			name:     "failed job with extra payload fields",
			key:      tenant.Production(),
			fixtures: []string{extraFields},
			want:     email.Failed{Error: "handler_error"},
		},
		{
			name:     "latest failed replay with extra payload fields",
			key:      tenant.Production(),
			fixtures: []string{failSend, failExtraReplay},
			want:     email.Failed{Error: "rejected"},
		},
		{
			name:     "pending job with extra payload fields",
			key:      tenant.Production(),
			fixtures: []string{failSend, pendingExtraFields},
			want:     email.Pending{},
		},
		{
			name:     "completed job",
			key:      tenant.Production(),
			fixtures: []string{failSend, completedSend},
			want:     email.Pending{},
		},
		{
			name:     "test tenant with delivery enabled",
			key:      branch,
			fixtures: []string{failSend},
			want:     email.Held{},
		},
	}
	ids := make([]uuid.UUID, 0, len(cases))
	stillPending := 0
	for _, tc := range cases {
		if tc.want == (email.Pending{}) {
			stillPending++
		}
		ctx := tenant.WithKey(t.Context(), tc.key)
		id := writeEmail(t, ctx, service, recoveryMail())
		ids = append(ids, id)
		for _, fixtureSQL := range tc.fixtures {
			if _, err := api.db.Pool.Exec(ctx, fixtureSQL, tc.key.String(), id.String()); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		}
	}

	scope, _ := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
	application := emailWorker(t, service, jobqueue.NewRepository(api.db.AppPool), scope)
	workerCtx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- application.Run(workerCtx) }()

	deadline := time.Now().Add(20 * time.Second)
	for {
		var pending int
		err := api.db.Pool.QueryRow(t.Context(), "select count(*) from emails where status = 'pending'").Scan(&pending)
		if err != nil {
			t.Fatal(err)
		}
		if pending == stillPending || time.Now().After(deadline) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Error(err)
	}

	for i, tc := range cases {
		row, err := service.Find(tenant.WithKey(t.Context(), tc.key), ids[i])
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if row.Delivery != tc.want {
			t.Errorf("%s: delivery=%+v want %+v", tc.name, row.Delivery, tc.want)
		}
	}

	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if fixture.attempts != 0 {
		t.Errorf("maintenance made %d SMTP attempts", fixture.attempts)
	}
}

func emailWorker(
	t *testing.T,
	service *email.Service,
	repository *jobqueue.Repository,
	scope jobqueue.Scope,
) *worker.Application {
	t.Helper()
	deps := workerDependencies(t, api.db.AppPool)
	deps.JobQueue = jobqueue.NewService(repository)
	deps.Email = service
	application, err := worker.NewApplication(deps, worker.Config{
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
		ownerPool, err := pgxpool.NewWithConfig(t.Context(), api.db.Pool.Config())
		if err != nil {
			t.Fatal(err)
		}
		defer ownerPool.Close()

		pause := &emailClaimPause{started: make(chan struct{}), resume: make(chan struct{})}
		config := api.db.AppPool.Config()
		config.ConnConfig.Tracer = pause
		pool, err := pgxpool.NewWithConfig(t.Context(), config)
		if err != nil {
			t.Fatal(err)
		}
		defer pool.Close()

		ctx := tenant.WithKey(t.Context(), tenant.Production())
		service := email.NewService(email.NewRepository(pool), nil)
		startupID := writeEmailWithPool(t, ctx, pool, service, recoveryMail())
		dailyID := writeEmailWithPool(t, ctx, pool, service, recoveryMail())
		for id, age := range map[uuid.UUID]time.Duration{
			startupID: 4 * 24 * time.Hour,
			dailyID:   2*24*time.Hour + time.Hour,
		} {
			_, err := ownerPool.Exec(ctx, `update emails set created_at = $2, status = 'held' where id = $1`,
				id, timex.Now().Add(-age))
			if err != nil {
				t.Fatal(err)
			}
		}

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
			err := ownerPool.QueryRow(ctx, `select
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
