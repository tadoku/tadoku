package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func workerTestDatabase(t *testing.T) *testpostgres.Database {
	t.Helper()
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	app := db.AppPool.Config().ConnConfig
	t.Setenv("WORKER_POSTGRES_HOST", app.Host)
	t.Setenv("WORKER_POSTGRES_PORT", fmt.Sprint(app.Port))
	t.Setenv("WORKER_POSTGRES_DATABASE", app.Database)
	t.Setenv("WORKER_POSTGRES_USER", app.User)
	t.Setenv("WORKER_POSTGRES_PASSWORD", app.Password)
	t.Setenv("WORKER_POSTGRES_SSLMODE", "disable")
	return db
}

func TestReplayCommandCreatesLinkedJob(t *testing.T) {
	db := workerTestDatabase(t)

	var failedID int64
	err := db.Pool.QueryRow(t.Context(), `insert into jobs (tenant, task_type, payload, state, attempts, failed_at, last_error)
		values ('tadoku/prod', 'leaderboard.invalidate_official.v1', '{"year":2025}', 'failed', 5, now(), 'handler_error') returning id`).Scan(&failedID)
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	args := []string{
		"--tenant", "tadoku/prod",
		"--id", fmt.Sprint(failedID),
		"--actor", "operator@example.test",
		"--reason", "provider restored",
	}
	if err := replay(t.Context(), args, logger); err != nil {
		t.Fatal(err)
	}

	var originalState, replayState, actor, reason string
	err = db.Pool.QueryRow(t.Context(), `select original.state, replay.state, replay.replay_actor, replay.replay_reason
		from jobs original join jobs replay on replay.replay_of_id = original.id where original.id = $1`, failedID).
		Scan(&originalState, &replayState, &actor, &reason)
	if err != nil {
		t.Fatal(err)
	}

	if originalState != "failed" || replayState != "pending" || actor != "operator@example.test" || reason != "provider restored" {
		t.Errorf("replay state: original=%q replay=%q actor=%q reason=%q", originalState, replayState, actor, reason)
	}

	missingActor := []string{"--tenant", "tadoku/prod", "--id", fmt.Sprint(failedID), "--reason", "missing actor"}
	if err := replay(context.Background(), missingActor, logger); err == nil {
		t.Error("replay accepted a missing actor")
	}
}

func TestReplayCommandReplaysOnlyInSelectedTenant(t *testing.T) {
	db := workerTestDatabase(t)

	const (
		jobTenant    = "e2e/replay-0123abcd"
		otherTenant  = "e2e/other-0123abcd"
		absentTenant = "e2e/absent-0123abcd"
	)
	_, err := db.Pool.Exec(
		t.Context(),
		"insert into tenants (key, kind) values ($1, 'test'), ($2, 'test')",
		jobTenant,
		otherTenant,
	)
	if err != nil {
		t.Fatal(err)
	}

	var failedID int64
	err = db.Pool.QueryRow(
		t.Context(),
		`insert into jobs (tenant, task_type, payload, state, attempts, failed_at, last_error)
		values ($1, 'leaderboard.invalidate_official.v1', '{"year":2025}', 'failed', 5, now(), 'handler_error')
		returning id`,
		jobTenant,
	).Scan(&failedID)
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	jobArgs := []string{
		"--id", fmt.Sprint(failedID),
		"--actor", "  operator@example.test ",
		"--reason", " provider restored  ",
	}
	for _, tc := range []struct {
		name   string
		tenant []string
		want   string
	}{
		{"missing tenant", nil, "--tenant: tenant key"},
		{"malformed tenant", []string{"--tenant", "Tadoku/Prod"}, "--tenant: tenant key"},
		{"production tenant", []string{"--tenant", "tadoku/prod"}, "tenant tadoku/prod: job is not failed"},
		{"other test tenant", []string{"--tenant", otherTenant}, "tenant " + otherTenant + ": job is not failed"},
		{"unregistered tenant", []string{"--tenant", absentTenant}, "tenant " + absentTenant + ": job is not failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := replay(t.Context(), append(tc.tenant, jobArgs...), logger)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("replay error = %v, want %q", err, tc.want)
			}

			var jobs int
			err = db.Pool.QueryRow(t.Context(), "select count(*) from jobs").Scan(&jobs)
			if err != nil || jobs != 1 {
				t.Fatalf("rejected replay changed jobs: count=%d error=%v", jobs, err)
			}
		})
	}

	if err := replay(t.Context(), append([]string{"--tenant", jobTenant}, jobArgs...), logger); err != nil {
		t.Fatal(err)
	}

	var originalState, replayTenant, replayState, taskType, payload, actor, reason string
	err = db.Pool.QueryRow(t.Context(), `select original.state, replay.tenant, replay.state, replay.task_type,
			replay.payload::text, replay.replay_actor, replay.replay_reason
		from jobs original join jobs replay on replay.replay_of_id = original.id
		where original.id = $1`,
		failedID,
	).Scan(&originalState, &replayTenant, &replayState, &taskType, &payload, &actor, &reason)
	if err != nil {
		t.Fatal(err)
	}

	got := []string{originalState, replayTenant, replayState, taskType, payload, actor, reason}
	want := []string{
		"failed",
		jobTenant,
		"pending",
		"leaderboard.invalidate_official.v1",
		`{"year": 2025}`,
		"operator@example.test",
		"provider restored",
	}
	if !slices.Equal(got, want) {
		t.Errorf("replayed job = %q, want %q", got, want)
	}
}

func TestLoadConfigRejectsInvalidWorkerSettings(t *testing.T) {
	t.Setenv("WORKER_BRANCH", "")
	t.Setenv("WORKER_COMPONENT", "tadoku-worker")
	t.Setenv("WORKER_POSTGRES_HOST", "127.0.0.1")
	t.Setenv("WORKER_POSTGRES_DATABASE", "postgres")
	t.Setenv("WORKER_POSTGRES_USER", "postgres")
	t.Setenv("WORKER_POSTGRES_PASSWORD", "postgres")
	t.Setenv("WORKER_POSTGRES_SSLMODE", "disable")

	for _, tc := range []struct {
		name  string
		key   string
		value string
		field string
	}{
		{"zero concurrency", "WORKER_CONCURRENCY", "0", "Concurrency"},
		{"negative concurrency", "WORKER_CONCURRENCY", "-1", "Concurrency"},
		{"malformed concurrency", "WORKER_CONCURRENCY", "invalid", "WORKER_CONCURRENCY"},
		{"zero shutdown timeout", "WORKER_SHUTDOWN_TIMEOUT", "0s", "ShutdownTimeout"},
		{"negative shutdown timeout", "WORKER_SHUTDOWN_TIMEOUT", "-1s", "ShutdownTimeout"},
		{"malformed shutdown timeout", "WORKER_SHUTDOWN_TIMEOUT", "invalid", "WORKER_SHUTDOWN_TIMEOUT"},
		{"invalid component", "WORKER_COMPONENT", "bad/component", "WORKER_COMPONENT"},
		{"empty component", "WORKER_COMPONENT", "", "WORKER_COMPONENT"},
		{"invalid branch", "WORKER_BRANCH", "INVALID", "WORKER_BRANCH"},
		{"unqualified branch", "WORKER_BRANCH", "tadoku", "WORKER_BRANCH"},
		{"production branch", "WORKER_BRANCH", "tadoku/prod", "WORKER_BRANCH"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			_, err := loadConfig()
			if err == nil || !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("loadConfig() error = %v, want rejection of %s", err, tc.key)
			}
		})
	}
}

func TestLoadConfigWorkerSettings(t *testing.T) {
	t.Setenv("WORKER_BRANCH", "")
	t.Setenv("WORKER_COMPONENT", "tadoku-worker")
	t.Setenv("WORKER_POSTGRES_HOST", "127.0.0.1")
	t.Setenv("WORKER_POSTGRES_DATABASE", "postgres")
	t.Setenv("WORKER_POSTGRES_USER", "postgres")
	t.Setenv("WORKER_POSTGRES_PASSWORD", "postgres")
	t.Setenv("WORKER_POSTGRES_SSLMODE", "disable")

	t.Run("defaults", func(t *testing.T) {
		t.Setenv("WORKER_COMPONENT", "")
		if err := os.Unsetenv("WORKER_COMPONENT"); err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Component != "tadoku-worker" {
			t.Errorf("default component=%q; want tadoku-worker", cfg.Component)
		}
		if cfg.Concurrency != 4 || cfg.ShutdownTimeout != 15*time.Second {
			t.Errorf("worker settings = (%d, %s), want (4, 15s)", cfg.Concurrency, cfg.ShutdownTimeout)
		}
	})

	t.Run("explicit settings", func(t *testing.T) {
		t.Setenv("WORKER_COMPONENT", "leaderboard-worker")
		t.Setenv("WORKER_CONCURRENCY", "7")
		t.Setenv("WORKER_SHUTDOWN_TIMEOUT", "3s")
		cfg, err := loadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Component != "leaderboard-worker" {
			t.Errorf("explicit component=%q; want leaderboard-worker", cfg.Component)
		}
		if cfg.Concurrency != 7 || cfg.ShutdownTimeout != 3*time.Second {
			t.Errorf("worker settings = (%d, %s), want (7, 3s)", cfg.Concurrency, cfg.ShutdownTimeout)
		}
	})

	t.Run("valid branch", func(t *testing.T) {
		t.Setenv("WORKER_BRANCH", "e2e/worker-0123abcd")
		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("loadConfig rejected a valid branch: %v", err)
		}
		if _, err := cfg.scope.Context(t.Context()); err != nil {
			t.Fatalf("valid branch scope rejected: %v", err)
		}
	})
}
