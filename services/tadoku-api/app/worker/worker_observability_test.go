package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/observability"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type jobTenantHandler struct {
	slog.Handler
	key tenant.Key
	ok  bool
}

func (handler *jobTenantHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Message == "job failed" {
		handler.key, handler.ok = tenant.FromContext(ctx)
	}
	return handler.Handler.Handle(ctx, record)
}

func TestWorkerObservationUsesPersistedTenant(t *testing.T) {
	fixture, err := newWorkerFixture(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	testKey, err := tenant.Parse("e2e/observation-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fixture.db.Exec(t.Context(), `insert into tenants (key, kind) values ($1, 'test')`, testKey.String())
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		key  tenant.Key
		kind string
	}{
		{tenant.Production(), "production"},
		{testKey, "test"},
	} {
		t.Run(test.kind, func(t *testing.T) {
			var id int64
			err := fixture.db.QueryRow(t.Context(), `
				insert into jobs (tenant, task_type, payload)
				values ($1, $2, '{"year":2026}')
				returning id`,
				test.key.String(),
				string(jobs.LeaderboardInvalidateOfficialV1),
			).Scan(&id)
			if err != nil {
				t.Fatal(err)
			}
			queue := jobqueue.NewService(jobqueue.NewRepository(fixture.database.AppPool))
			claimed, err := queue.Claim(
				tenant.WithKey(t.Context(), test.key),
				jobs.LeaderboardInvalidateOfficialV1,
				1,
				time.Minute,
				1,
			)
			if err != nil || len(claimed) != 1 || claimed[0].ID != id {
				t.Fatalf("claim fixture: count=%d error=%v", len(claimed), err)
			}

			failure := &permanentError{err: errors.New("observation fixture failure")}
			handlers, err := newRegistry(handle(
				func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error { return failure },
				Policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 1},
			))
			if err != nil {
				t.Fatal(err)
			}
			registry := prometheus.NewRegistry()
			var logs bytes.Buffer
			captured := &jobTenantHandler{Handler: observability.NewTenantHandler(slog.NewJSONHandler(&logs, nil))}
			runtime := &runner{
				queue:    queue,
				handlers: handlers,
				logger:   slog.New(captured),
				metrics:  NewMetrics(registry),
			}

			err = runtime.process(t.Context(), claimed[0], handlers.ordered[0].spec)
			if !errors.Is(err, failure) {
				t.Fatalf("process failure=%v, want fixture failure", err)
			}
			if !captured.ok || captured.key != test.key {
				t.Errorf("job failure context tenant=%q known=%t, want %q", captured.key.String(), captured.ok, test.key)
			}
			var event map[string]any
			if err := json.Unmarshal(logs.Bytes(), &event); err != nil {
				t.Fatal(err)
			}
			if event["msg"] != "job failed" || event["tenant"] != test.key.String() {
				t.Errorf("job failure event=%v, want persisted tenant %q", event, test.key)
			}
			var state string
			err = fixture.db.QueryRow(t.Context(), `select state from jobs where id = $1`, id).Scan(&state)
			if err != nil || state != "failed" {
				t.Errorf("persisted failure state=%q error=%v", state, err)
			}

			families, err := registry.Gather()
			if err != nil {
				t.Fatal(err)
			}
			observed := 0
			for _, family := range families {
				name := family.GetName()
				if name != "tadoku_worker_failed_attempts_total" && name != "tadoku_worker_handler_duration_seconds" {
					continue
				}
				for _, metric := range family.GetMetric() {
					labels := make(map[string]string)
					for _, label := range metric.GetLabel() {
						labels[label.GetName()] = label.GetValue()
					}
					if labels["type"] != string(jobs.LeaderboardInvalidateOfficialV1) || labels["tenant_kind"] != test.kind {
						t.Errorf("%s labels=%v, want persisted tenant_kind=%q and existing type", name, labels, test.kind)
					}
					if name == "tadoku_worker_failed_attempts_total" {
						if labels["code"] != failureCode(failure) || len(labels) != 3 || metric.GetCounter().GetValue() != 1 {
							t.Errorf("failed-attempt sample labels=%v count=%v", labels, metric.GetCounter().GetValue())
						}
					} else if len(labels) != 2 || metric.GetHistogram().GetSampleCount() != 1 {
						t.Errorf("handler-duration sample labels=%v count=%d", labels, metric.GetHistogram().GetSampleCount())
					}
					observed++
				}
			}
			if observed != 2 {
				t.Errorf("observed %d job metric samples, want 2", observed)
			}
		})
	}
}
