package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestLeaderboardOutboxIsolatesBranchCaches(t *testing.T) {
	api.reset(t, "testdata/ImmersionFetchLeaderboardGlobal/200_cache_miss")
	branchDB, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := branchDB.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := branchDB.Reset(t.Context(), "testdata/ImmersionFetchLeaderboardGlobal/200_cache_miss/setup.sql"); err != nil {
		t.Fatal(err)
	}

	basePrefix := "dev:base-" + uuid.NewString() + ":"
	branchPrefix := "dev:branch-" + uuid.NewString() + ":"
	baseKey := basePrefix + "leaderboard:global"
	branchKey := branchPrefix + "leaderboard:global"
	keys := []string{
		baseKey, baseKey + ":last_updated", baseKey + ":generation",
		branchKey, branchKey + ":last_updated", branchKey + ":generation",
		branchPrefix + "leaderboard:yearly:2026", branchPrefix + "leaderboard:yearly:2026:last_updated", branchPrefix + "leaderboard:yearly:2026:generation",
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := leaderboardValkey.client.Do(ctx, leaderboardValkey.client.B().Del().Key(keys...).Build()).Error(); err != nil {
			t.Error(err)
		}
	})

	base := leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second, basePrefix)
	branch := leaderboard.NewService(leaderboard.NewRepository(branchDB.Pool), leaderboardValkey.client, time.Second, branchPrefix)
	for _, service := range []*leaderboard.Service{base, branch} {
		if _, err := service.FetchGlobal(t.Context(), leaderboard.Request{}); err != nil {
			t.Fatal(err)
		}
	}
	baseBefore := cachedLeaderboardScore(t, baseKey)
	if got := cachedLeaderboardScore(t, branchKey); got != baseBefore {
		t.Fatalf("initial scores differ across identical database fixtures: base=%v branch=%v", baseBefore, got)
	}
	worker := leaderboard.NewWorker(branch, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := worker.ProcessPending(t.Context()); err != nil {
		t.Fatal(err)
	}
	baseMarker, err := leaderboardValkey.client.Do(t.Context(), leaderboardValkey.client.B().Exists().Key(baseKey+":last_updated").Build()).ToInt64()
	if err != nil {
		t.Fatal(err)
	}
	branchMarker, err := leaderboardValkey.client.Do(t.Context(), leaderboardValkey.client.B().Exists().Key(branchKey+":last_updated").Build()).ToInt64()
	if err != nil {
		t.Fatal(err)
	}
	if baseMarker != 1 || branchMarker != 0 {
		t.Fatalf("startup scan markers: base=%d branch=%d, want 1 and 0", baseMarker, branchMarker)
	}
	if _, err := branch.FetchGlobal(t.Context(), leaderboard.Request{}); err != nil {
		t.Fatal(err)
	}

	if _, err := branchDB.Pool.Exec(t.Context(), `update logs set amount = 100 where id = 'a1111111-1111-4111-8111-000000000001'`); err != nil {
		t.Fatal(err)
	}
	if _, err := branchDB.Pool.Exec(t.Context(), `insert into leaderboard_outbox (event_type, user_id, year) values ('refresh_official_scores', '11111111-1111-4111-8111-111111111111', 2026)`); err != nil {
		t.Fatal(err)
	}
	if err := worker.ProcessPending(t.Context()); err != nil {
		t.Fatal(err)
	}
	var pending int
	if err := branchDB.Pool.QueryRow(t.Context(), `select count(*) from leaderboard_outbox where processed_at is null`).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	if pending != 0 {
		t.Errorf("branch outbox has %d pending events after worker pass", pending)
	}

	baseMarker, err = leaderboardValkey.client.Do(t.Context(), leaderboardValkey.client.B().Exists().Key(baseKey+":last_updated").Build()).ToInt64()
	if err != nil {
		t.Fatal(err)
	}
	if baseMarker != 1 || cachedLeaderboardScore(t, baseKey) != baseBefore {
		t.Error("branch worker changed the base leaderboard cache")
	}
	branchMarker, err = leaderboardValkey.client.Do(t.Context(), leaderboardValkey.client.B().Exists().Key(branchKey+":last_updated").Build()).ToInt64()
	if err != nil {
		t.Fatal(err)
	}
	if branchMarker != 0 {
		t.Error("branch worker did not invalidate its own cache")
	}
	if _, err := branch.FetchGlobal(t.Context(), leaderboard.Request{}); err != nil {
		t.Fatal(err)
	}
	if got := cachedLeaderboardScore(t, branchKey); got <= baseBefore {
		t.Errorf("branch rebuilt score = %v, want greater than unchanged base score %v", got, baseBefore)
	}
}

func cachedLeaderboardScore(t *testing.T, key string) float64 {
	t.Helper()
	entries, err := leaderboardValkey.client.Do(t.Context(), leaderboardValkey.client.B().Zrange().Key(key).Min("0").Max("-1").Withscores().Build()).AsZScores()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Member == "11111111-1111-4111-8111-111111111111" {
			return entry.Score
		}
	}
	t.Fatalf("user score absent from cache %s", key)
	return 0
}

func TestLeaderboardOutboxStartupReconcilesWarmCaches(t *testing.T) {
	api.reset(t, "testdata/ImmersionFetchLeaderboardGlobal/200_cache_miss")
	keys := []string{
		"leaderboard:global",
		"leaderboard:yearly:2026",
		"leaderboard:contest:f1111111-1111-4111-8111-111111111111",
	}
	for _, key := range keys {
		seedLeaderboardCache(t, key, "hit")
	}

	worker := leaderboard.NewWorker(leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second, ""), slog.New(slog.NewTextHandler(io.Discard, nil)))
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
	api.reset(t, "testdata/ImmersionFetchLeaderboardGlobal/200_cache_miss")
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
	failedWorker := leaderboard.NewWorker(leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), closed, 25*time.Millisecond, ""), slog.New(slog.NewTextHandler(io.Discard, nil)))
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

	worker := leaderboard.NewWorker(leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second, ""), slog.New(slog.NewTextHandler(io.Discard, nil)))
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
	api.reset(t, "testdata/ImmersionFetchLeaderboardGlobal/200_cache_miss")
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
	worker := leaderboard.NewWorker(leaderboard.NewService(leaderboard.NewRepository(api.db.Pool), leaderboardValkey.client, time.Second, ""), slog.New(slog.NewTextHandler(io.Discard, nil)))
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
