package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestReplayCommandCreatesLinkedTask(t *testing.T) {
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	parsed, err := url.Parse(db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	password, _ := parsed.User.Password()
	t.Setenv("WORKER_POSTGRES_HOST", parsed.Hostname())
	t.Setenv("WORKER_POSTGRES_PORT", parsed.Port())
	t.Setenv("WORKER_POSTGRES_DATABASE", parsed.Path[1:])
	t.Setenv("WORKER_POSTGRES_USER", parsed.User.Username())
	t.Setenv("WORKER_POSTGRES_PASSWORD", password)
	t.Setenv("WORKER_POSTGRES_SSLMODE", "disable")

	var failedID int64
	err = db.Pool.QueryRow(t.Context(), `insert into jobs (task_type, payload, state, attempts, failed_at, last_error)
		values ('leaderboard.invalidate_official.v1', '{"year":2025}', 'failed', 5, now(), 'handler_error') returning id`).Scan(&failedID)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := replay(t.Context(), []string{"--id", fmt.Sprint(failedID), "--actor", "operator@example.test", "--reason", "provider restored"}, logger); err != nil {
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
	if err := replay(context.Background(), []string{"--id", fmt.Sprint(failedID), "--reason", "missing actor"}, logger); err == nil {
		t.Error("replay accepted a missing actor")
	}
}

func TestLoadConfigRejectsInvalidWorkerSettings(t *testing.T) {
	t.Setenv("WORKER_VALKEY_URL", "redis://127.0.0.1:6379")
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
	t.Setenv("WORKER_VALKEY_URL", "redis://127.0.0.1:6379")
	t.Setenv("WORKER_POSTGRES_HOST", "127.0.0.1")
	t.Setenv("WORKER_POSTGRES_DATABASE", "postgres")
	t.Setenv("WORKER_POSTGRES_USER", "postgres")
	t.Setenv("WORKER_POSTGRES_PASSWORD", "postgres")
	t.Setenv("WORKER_POSTGRES_SSLMODE", "disable")

	t.Run("defaults", func(t *testing.T) {
		cfg, err := loadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Concurrency != 4 || cfg.ShutdownTimeout != 15*time.Second {
			t.Errorf("worker settings = (%d, %s), want (4, 15s)", cfg.Concurrency, cfg.ShutdownTimeout)
		}
	})
	t.Run("explicit settings", func(t *testing.T) {
		t.Setenv("WORKER_CONCURRENCY", "7")
		t.Setenv("WORKER_SHUTDOWN_TIMEOUT", "3s")
		cfg, err := loadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Concurrency != 7 || cfg.ShutdownTimeout != 3*time.Second {
			t.Errorf("worker settings = (%d, %s), want (7, 3s)", cfg.Concurrency, cfg.ShutdownTimeout)
		}
	})
}
