package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
)

func TestLeaderboardOutboxStartupReconcilesWarmCaches(t *testing.T) {
	api.reset(t, "testdata/ImmersionFetchLeaderboardGlobal/200_cache_hit")
	keys := []string{
		"leaderboard:global",
		"leaderboard:yearly:2026",
		"leaderboard:contest:f1111111-1111-4111-8111-111111111111",
	}
	for _, key := range keys {
		seedLeaderboardCache(t, key, "hit")
	}

	worker := leaderboard.NewWorker(leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		worker.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})

	deadline := time.After(3 * time.Second)
	for {
		reconciled := true
		for _, key := range keys {
			generation, err := leaderboardValkey.client.Do(t.Context(), leaderboardValkey.client.B().Get().Key(key+":generation").Build()).ToString()
			if err != nil || generation == "0" {
				reconciled = false
				break
			}
		}
		if reconciled {
			break
		}
		select {
		case <-deadline:
			t.Fatal("startup reconciliation did not invalidate warm caches")
		case <-time.After(10 * time.Millisecond):
		}
	}

}

func TestLeaderboardOutboxRetriesAfterValkeyFailure(t *testing.T) {
	api.reset(t, "testdata/ImmersionFetchLeaderboardGlobal/200_cache_hit")
	seedLeaderboardCache(t, "leaderboard:global", "hit")

	if _, err := api.db.Pool.Exec(t.Context(), `update logs set amount = 20 where id = 'a1111111-1111-4111-8111-000000000001'`); err != nil {
		t.Fatal(err)
	}
	if _, err := api.db.Pool.Exec(t.Context(), `insert into leaderboard_outbox (event_type, user_id, year) values ('refresh_official_scores', '11111111-1111-4111-8111-111111111111', 2026)`); err != nil {
		t.Fatal(err)
	}

	closed, err := newClosedLeaderboardValkeyClient()
	if err != nil {
		t.Fatal(err)
	}
	failedWorker := leaderboard.NewWorker(leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), closed, 25*time.Millisecond), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if _, err := failedWorker.ProcessBatch(t.Context()); err == nil {
		t.Fatal("outbox batch succeeded with unavailable Valkey")
	}
	var pending int
	if err := api.db.Pool.QueryRow(t.Context(), `select count(*) from leaderboard_outbox where processed_at is null`).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	if pending != 1 {
		t.Fatalf("pending events after Valkey failure = %d, want 1", pending)
	}

	worker := leaderboard.NewWorker(leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)))
	processed, err := worker.ProcessBatch(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if processed != 1 {
		t.Fatalf("processed events = %d, want 1", processed)
	}
	if err := api.db.Pool.QueryRow(t.Context(), `select count(*) from leaderboard_outbox where processed_at is null`).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	if pending != 0 {
		t.Errorf("pending events after recovery = %d, want 0", pending)
	}

	markers, err := leaderboardValkey.client.Do(t.Context(), leaderboardValkey.client.B().Exists().Key("leaderboard:global:last_updated").Build()).ToInt64()
	if err != nil {
		t.Fatal(err)
	}
	if markers != 0 {
		t.Errorf("global cache marker after recovered outbox batch = %d, want 0", markers)
	}
}

func TestLeaderboardOutboxSkipsLockedEventsAndRetries(t *testing.T) {
	api.reset(t, "testdata/ImmersionFetchLeaderboardGlobal/200_cache_hit")
	if _, err := api.db.Pool.Exec(t.Context(), `insert into leaderboard_outbox (event_type, user_id, year) values ('refresh_official_scores', '11111111-1111-4111-8111-111111111111', 2026)`); err != nil {
		t.Fatal(err)
	}
	tx, err := api.db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(t.Context())
	if _, err := tx.Exec(t.Context(), `select id from leaderboard_outbox where processed_at is null for update`); err != nil {
		t.Fatal(err)
	}
	worker := leaderboard.NewWorker(leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)))
	processed, err := worker.ProcessBatch(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if processed != 0 {
		t.Errorf("processed locked event = %d, want 0", processed)
	}
	if err := tx.Rollback(t.Context()); err != nil {
		t.Fatal(err)
	}
	processed, err = worker.ProcessBatch(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if processed != 1 {
		t.Errorf("processed after lock release = %d, want 1", processed)
	}
}
