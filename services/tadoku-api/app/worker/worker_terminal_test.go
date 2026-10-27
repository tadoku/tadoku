package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestWorkerCountsOnlyTerminalFailures(t *testing.T) {
	f, err := newWorkerFixture(tenant.WithKey(t.Context(), tenant.Production()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})

	var mu sync.Mutex
	calls := map[int16]int{}
	handlers, err := newRegistry(handle(func(_ context.Context, job jobs.InvalidateOfficialLeaderboardV1) error {
		mu.Lock()
		defer mu.Unlock()
		calls[job.Year]++
		if job.Year == 2001 || calls[job.Year] == 1 {
			return errors.New("terminal fixture failure")
		}
		return nil
	}, policy{Concurrency: 4, Timeout: time.Second, MaxAttempts: 2}))
	if err != nil {
		t.Fatal(err)
	}
	registry := prometheus.NewRegistry()
	deps := workerDependencies(f.db, nil)
	runtime := &runner{
		scope:           baseWorkerScope(t),
		queue:           deps.JobQueue,
		email:           deps.Email,
		identity:        deps.IdentityExpiry,
		handlers:        handlers,
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		metrics:         NewMetrics(registry),
		shutdownTimeout: 2 * time.Second,
		concurrency:     4,
	}
	runtime.metrics.initialize(handlers)

	production, test := tenant.Production().String(), f.key.String()
	official := string(jobs.LeaderboardInvalidateOfficialV1)
	ids := map[string]int64{}
	for _, row := range []struct {
		name, tenant, payload, state string
		attempts                     int
		expired                      bool
	}{
		{"exhausted", production, `{"year":2001}`, "pending", 0, false},
		{"recovered", production, `{"year":2002}`, "pending", 0, false},
		{"invalid", production, `{"year":0}`, "pending", 0, false},
		{"test-exhausted", test, `{"year":2001}`, "pending", 0, false},
		{"lease-expired", production, `{"year":2003}`, "running", 2, true},
		{"limit-lowered", production, `{"year":2004}`, "pending", 2, false},
	} {
		var id int64
		err := f.db.QueryRow(
			t.Context(),
			`insert into jobs (tenant, task_type, payload, state, attempts, claim_token, lease_expires_at)
			values ($1, $2, $3::jsonb, $4, $5,
				case when $6 then $7::uuid end,
				case when $6 then now() - interval '1 second' end)
			returning id`,
			row.tenant,
			official,
			row.payload,
			row.state,
			row.attempts,
			row.expired,
			uuid.New(),
		).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		ids[row.name] = id
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); runtime.run(ctx) }()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	want := map[string]string{
		"exhausted":      "failed",
		"recovered":      "completed",
		"invalid":        "failed",
		"test-exhausted": "failed",
		"lease-expired":  "failed",
		"limit-lowered":  "failed",
	}
	err = waitFor(tenant.WithKey(t.Context(), tenant.Production()), func() (bool, error) {
		for name, state := range want {
			var got string
			err := f.db.QueryRow(t.Context(), `select state from jobs where id = $1`, ids[name]).Scan(&got)
			if err != nil || got != state {
				return false, err
			}
		}
		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	<-done

	terminal := map[string]float64{}
	for _, code := range []string{
		"invalid_payload",
		"deadline_exceeded",
		"handler_error",
		"lease_expired",
		"attempts_exhausted",
	} {
		for _, kind := range []string{"production", "test", "unknown"} {
			terminal[official+"|"+code+"|"+kind] = 0
		}
	}
	terminal[official+"|handler_error|production"] = 1
	terminal[official+"|invalid_payload|production"] = 1
	terminal[official+"|handler_error|test"] = 1
	terminal[official+"|lease_expired|production"] = 1
	terminal[official+"|attempts_exhausted|production"] = 1
	assertCounters(t, registry, "tadoku_worker_terminal_failures_total", terminal)
	assertCounters(t, registry, "tadoku_worker_failed_attempts_total", map[string]float64{
		official + "|handler_error|production":   3,
		official + "|invalid_payload|production": 1,
		official + "|handler_error|test":         2,
	})
}

func assertCounters(t *testing.T, registry *prometheus.Registry, name string, want map[string]float64) {
	t.Helper()

	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]float64{}
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := map[string]string{}
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			got[labels["type"]+"|"+labels["code"]+"|"+labels["tenant_kind"]] = metric.GetCounter().GetValue()
		}
	}

	if len(got) != len(want) {
		t.Errorf("%s series=%v, want %v", name, got, want)
		return
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s{%s}=%v, want %v; all=%v", name, key, got[key], value, got)
		}
	}
}
