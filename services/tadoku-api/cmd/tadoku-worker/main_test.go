package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"testing"

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
	err = db.Pool.QueryRow(t.Context(), `insert into async_outbox (task_type, payload, state, attempts, failed_at, last_error)
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
		from async_outbox original join async_outbox replay on replay.replay_of_id = original.id where original.id = $1`, failedID).
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
