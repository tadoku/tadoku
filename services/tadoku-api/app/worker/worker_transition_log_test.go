package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestWorkerLogsTransitionFailureAfterSuccessfulHandler(t *testing.T) {
	fixture, err := newWorkerFixture(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	pool, err := pgxpool.New(t.Context(), fixture.dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	var id int64
	err = fixture.db.QueryRow(t.Context(), `
		insert into jobs (tenant, task_type, payload)
		values ($1, $2, '{"year":2026}')
		returning id`,
		tenant.Production().String(),
		string(jobs.LeaderboardInvalidateOfficialV1),
	).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	queue := jobqueue.NewService(jobqueue.NewRepository(pool))
	claimed, err := queue.Claim(
		tenant.WithKey(t.Context(), tenant.Production()),
		jobs.LeaderboardInvalidateOfficialV1,
		1,
		time.Minute,
		3,
	)
	if err != nil || len(claimed) != 1 || claimed[0].ID != id {
		t.Fatalf("claim fixture: count=%d error=%v", len(claimed), err)
	}

	handlers, err := newRegistry(handle(
		func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error {
			pool.Close()
			return nil
		},
		policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 3},
	))
	if err != nil {
		t.Fatal(err)
	}

	var logs bytes.Buffer
	runtime := &runner{
		queue:    queue,
		handlers: handlers,
		logger:   slog.New(slog.NewJSONHandler(&logs, nil)),
		metrics:  NewMetrics(prometheus.NewRegistry()),
	}
	spec := handlers.ordered[0].spec
	spec.lease = time.Minute

	err = runtime.process(t.Context(), claimed[0], spec)
	if err == nil {
		t.Fatal("process error = nil, want transition failure")
	}

	events := decodeLogEvents(t, logs.Bytes())
	transition := eventsWithMessage(events, "transition job")
	if len(transition) != 1 {
		t.Fatalf("transition job events=%v, want exactly one", events)
	}
	if got, ok := transition[0]["job_id"].(float64); !ok || int64(got) != id {
		t.Errorf("transition job_id=%v, want %d", transition[0]["job_id"], id)
	}
	if transition[0]["type"] != string(jobs.LeaderboardInvalidateOfficialV1) {
		t.Errorf("transition type=%v, want %s", transition[0]["type"], jobs.LeaderboardInvalidateOfficialV1)
	}
	if transition[0]["level"] != "ERROR" {
		t.Errorf("transition level=%v, want ERROR", transition[0]["level"])
	}
	if len(eventsWithMessage(events, "job failed")) != 0 {
		t.Errorf("unexpected job failed events in %v", events)
	}
}

func TestWorkerLogsHandlerErrorWhenTransitionFails(t *testing.T) {
	fixture, err := newWorkerFixture(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	pool, err := pgxpool.New(t.Context(), fixture.dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	var id int64
	err = fixture.db.QueryRow(t.Context(), `
		insert into jobs (tenant, task_type, payload)
		values ($1, $2, '{"year":2026}')
		returning id`,
		tenant.Production().String(),
		string(jobs.LeaderboardInvalidateOfficialV1),
	).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	queue := jobqueue.NewService(jobqueue.NewRepository(pool))
	claimed, err := queue.Claim(
		tenant.WithKey(t.Context(), tenant.Production()),
		jobs.LeaderboardInvalidateOfficialV1,
		1,
		time.Minute,
		3,
	)
	if err != nil || len(claimed) != 1 || claimed[0].ID != id {
		t.Fatalf("claim fixture: count=%d error=%v", len(claimed), err)
	}

	failure := errors.New("handler fixture failure")
	handlers, err := newRegistry(handle(
		func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error {
			pool.Close()
			return failure
		},
		policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 3},
	))
	if err != nil {
		t.Fatal(err)
	}

	var logs bytes.Buffer
	runtime := &runner{
		queue:    queue,
		handlers: handlers,
		logger:   slog.New(slog.NewJSONHandler(&logs, nil)),
		metrics:  NewMetrics(prometheus.NewRegistry()),
	}
	spec := handlers.ordered[0].spec
	spec.lease = time.Minute

	err = runtime.process(t.Context(), claimed[0], spec)
	if err == nil {
		t.Fatal("process error = nil, want transition failure")
	}

	events := decodeLogEvents(t, logs.Bytes())
	transition := eventsWithMessage(events, "transition job")
	if len(transition) != 1 {
		t.Fatalf("transition job events=%v, want exactly one", events)
	}
	if transition[0]["handler_error"] != failure.Error() {
		t.Errorf("handler_error=%v, want %q", transition[0]["handler_error"], failure.Error())
	}
	if transition[0]["code"] != failureCode(failure) {
		t.Errorf("code=%v, want %s", transition[0]["code"], failureCode(failure))
	}
	if len(eventsWithMessage(events, "job failed")) != 0 {
		t.Errorf("unexpected job failed events in %v", events)
	}
}

func TestWorkerLogsHandlerFailureOnce(t *testing.T) {
	fixture, err := newWorkerFixture(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	var id int64
	err = fixture.db.QueryRow(t.Context(), `
		insert into jobs (tenant, task_type, payload)
		values ($1, $2, '{"year":2026}')
		returning id`,
		tenant.Production().String(),
		string(jobs.LeaderboardInvalidateOfficialV1),
	).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	queue := jobqueue.NewService(jobqueue.NewRepository(fixture.database.AppPool))
	claimed, err := queue.Claim(
		tenant.WithKey(t.Context(), tenant.Production()),
		jobs.LeaderboardInvalidateOfficialV1,
		1,
		time.Minute,
		3,
	)
	if err != nil || len(claimed) != 1 || claimed[0].ID != id {
		t.Fatalf("claim fixture: count=%d error=%v", len(claimed), err)
	}

	failure := errors.New("handler fixture failure")
	handlers, err := newRegistry(handle(
		func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error { return failure },
		policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 3},
	))
	if err != nil {
		t.Fatal(err)
	}

	var logs bytes.Buffer
	runtime := &runner{
		queue:    queue,
		handlers: handlers,
		logger:   slog.New(slog.NewJSONHandler(&logs, nil)),
		metrics:  NewMetrics(prometheus.NewRegistry()),
	}
	spec := handlers.ordered[0].spec
	spec.lease = time.Minute

	err = runtime.process(t.Context(), claimed[0], spec)
	if !errors.Is(err, failure) {
		t.Fatalf("process error=%v, want handler fixture failure", err)
	}

	events := decodeLogEvents(t, logs.Bytes())
	failed := eventsWithMessage(events, "job failed")
	if len(failed) != 1 {
		t.Fatalf("job failed events=%v, want exactly one", events)
	}
	if got, ok := failed[0]["job_id"].(float64); !ok || int64(got) != id {
		t.Errorf("job failed job_id=%v, want %d", failed[0]["job_id"], id)
	}
	if len(eventsWithMessage(events, "transition job")) != 0 {
		t.Errorf("unexpected transition job events in %v", events)
	}
}

func decodeLogEvents(t *testing.T, raw []byte) []map[string]any {
	t.Helper()

	var events []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatalf("decode log line %q: %v", line, err)
		}
		events = append(events, event)
	}
	return events
}

func eventsWithMessage(events []map[string]any, message string) []map[string]any {
	var matched []map[string]any
	for _, event := range events {
		if event["msg"] == message {
			matched = append(matched, event)
		}
	}
	return matched
}
