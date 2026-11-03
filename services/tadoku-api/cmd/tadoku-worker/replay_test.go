package main

import (
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func insertLineageJob(t *testing.T, db *testpostgres.Database, state string, replayOf int64) int64 {
	t.Helper()

	var failedAt, completedAt, source, actor, reason any
	switch state {
	case "failed":
		failedAt = time.Now()
	case "completed":
		completedAt = time.Now()
	}
	if replayOf != 0 {
		source, actor, reason = replayOf, "operator@example.test", "provider restored"
	}

	var id int64
	err := db.Pool.QueryRow(
		t.Context(),
		`insert into jobs (
			tenant, task_type, payload, state, attempts, failed_at, completed_at,
			replay_of_id, replay_actor, replay_reason
		)
		values ('tadoku/prod', 'leaderboard.invalidate_official.v1', '{"year":2025}', $1, 1, $2, $3, $4, $5, $6)
		returning id`,
		state,
		failedAt,
		completedAt,
		source,
		actor,
		reason,
	).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func replayJob(t *testing.T, id int64) error {
	t.Helper()
	args := []string{
		"--tenant", "tadoku/prod",
		"--id", fmt.Sprint(id),
		"--actor", "operator@example.test",
		"--reason", "provider restored",
	}
	return replay(t.Context(), args, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func countReplays(t *testing.T, db *testpostgres.Database, of int64) int {
	t.Helper()
	var count int
	err := db.Pool.QueryRow(t.Context(), `select count(*) from jobs where replay_of_id = $1`, of).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func wantRefused(t *testing.T, err error, id int64, want string) {
	t.Helper()
	wantMessage := fmt.Sprintf("replay job %d in tenant tadoku/prod: %s", id, want)
	if err == nil || err.Error() != wantMessage {
		t.Errorf("replay job %d error = %v, want %q", id, err, wantMessage)
	}
	if errx.KindOf(err) != errx.Conflict {
		t.Errorf("replay job %d error kind = %v, want conflict", id, errx.KindOf(err))
	}
}

func TestReplayCommandRefusesAJobThatWasAlreadyReplayed(t *testing.T) {
	db := workerTestDatabase(t)
	failedID := insertLineageJob(t, db, "failed", 0)
	replayID := insertLineageJob(t, db, "pending", failedID)

	err := replayJob(t, failedID)

	wantRefused(t, err, failedID, fmt.Sprintf("job was already replayed as job %d", replayID))
	if got := countReplays(t, db, failedID); got != 1 {
		t.Errorf("replays of job %d = %d, want 1", failedID, got)
	}
}

func TestReplayCommandInsertsOneReplayWhenTwoRunAtOnce(t *testing.T) {
	db := workerTestDatabase(t)
	failedID := insertLineageJob(t, db, "failed", 0)

	tx, err := db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(t.Context()) }()
	var holder int32
	var competingID int64
	err = tx.QueryRow(t.Context(), `select pg_backend_pid() from jobs where id = $1 for update`, failedID).Scan(&holder)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.QueryRow(
		t.Context(),
		`insert into jobs (tenant, task_type, payload, replay_of_id, replay_actor, replay_reason)
		select tenant, task_type, payload, id, 'other-operator', 'provider restored'
		from jobs
		where id = $1
		returning id`,
		failedID,
	).Scan(&competingID)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- replayJob(t, failedID) }()
	finished, result := waitForBlockedOrDone(t, db, holder, done)
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !finished {
		result = <-done
	}

	if got := countReplays(t, db, failedID); got != 1 {
		t.Errorf("replays of job %d after two concurrent replays = %d, want 1", failedID, got)
	}
	wantRefused(t, result, failedID, fmt.Sprintf("job was already replayed as job %d", competingID))
}

func waitForBlockedOrDone(
	t *testing.T,
	db *testpostgres.Database,
	holder int32,
	done <-chan error,
) (bool, error) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			return true, err
		default:
		}

		var blocked bool
		err := db.Pool.QueryRow(
			t.Context(),
			`select exists (select 1 from pg_stat_activity where $1 = any(pg_blocking_pids(pid)))`,
			holder,
		).Scan(&blocked)
		if err != nil {
			t.Fatal(err)
		}
		if blocked {
			return false, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("replay neither finished nor waited for the held row lock")
	return false, nil
}

func TestReplayCommandReplaysOnlyTheNewestJobOfALineage(t *testing.T) {
	db := workerTestDatabase(t)
	originalID := insertLineageJob(t, db, "failed", 0)
	failedReplayID := insertLineageJob(t, db, "failed", originalID)

	if err := replayJob(t, failedReplayID); err != nil {
		t.Fatalf("replay of the failed replay: %v", err)
	}
	var newestID int64
	err := db.Pool.QueryRow(t.Context(), `select id from jobs where replay_of_id = $1`, failedReplayID).Scan(&newestID)
	if err != nil {
		t.Fatal(err)
	}

	wantRefused(
		t,
		replayJob(t, originalID),
		originalID,
		fmt.Sprintf("job was already replayed as job %d", failedReplayID),
	)
	wantRefused(
		t,
		replayJob(t, failedReplayID),
		failedReplayID,
		fmt.Sprintf("job was already replayed as job %d", newestID),
	)
	if got := countReplays(t, db, originalID) + countReplays(t, db, failedReplayID); got != 2 {
		t.Errorf("replays in the lineage = %d, want 2", got)
	}
}

func TestReplayCommandRefusesALineageWhoseReplaySucceeded(t *testing.T) {
	db := workerTestDatabase(t)
	originalID := insertLineageJob(t, db, "failed", 0)
	failedReplayID := insertLineageJob(t, db, "failed", originalID)
	succeededID := insertLineageJob(t, db, "completed", failedReplayID)

	for _, tc := range []struct {
		id   int64
		want string
	}{
		{originalID, fmt.Sprintf("job was already replayed as job %d", failedReplayID)},
		{failedReplayID, fmt.Sprintf("job was already replayed as job %d", succeededID)},
		{succeededID, "job is not failed or supported"},
	} {
		wantRefused(t, replayJob(t, tc.id), tc.id, tc.want)
	}

	var jobs int
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from jobs`).Scan(&jobs); err != nil {
		t.Fatal(err)
	}
	if jobs != 3 {
		t.Errorf("jobs after refused replays = %d, want 3", jobs)
	}
}

func countJobs(t *testing.T, db *testpostgres.Database) int {
	t.Helper()
	var count int
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from jobs`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestReplayCommandRefusesAForkedReplayWhoseSiblingSucceeded(t *testing.T) {
	db := workerTestDatabase(t)
	originalID := insertLineageJob(t, db, "failed", 0)
	failedReplayID := insertLineageJob(t, db, "failed", originalID)
	succeededID := insertLineageJob(t, db, "completed", originalID)

	err := replayJob(t, failedReplayID)

	wantRefused(t, err, failedReplayID, fmt.Sprintf("job lineage cannot be replayed: job %d is completed", succeededID))
	if got := countJobs(t, db); got != 3 {
		t.Errorf("jobs after a refused replay of a forked lineage = %d, want 3", got)
	}
}

func TestReplayCommandInsertsOneReplayWhenForkedSiblingsRunAtOnce(t *testing.T) {
	db := workerTestDatabase(t)
	originalID := insertLineageJob(t, db, "failed", 0)
	firstID := insertLineageJob(t, db, "failed", originalID)
	secondID := insertLineageJob(t, db, "failed", originalID)

	tx, err := db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(t.Context()) }()
	var holder int32
	var competingID int64
	err = tx.QueryRow(t.Context(), `select pg_backend_pid() from jobs where id = $1 for update`, originalID).Scan(&holder)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.QueryRow(
		t.Context(),
		`insert into jobs (tenant, task_type, payload, replay_of_id, replay_actor, replay_reason)
		select tenant, task_type, payload, id, 'other-operator', 'provider restored'
		from jobs
		where id = $1
		returning id`,
		firstID,
	).Scan(&competingID)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- replayJob(t, secondID) }()
	finished, result := waitForBlockedOrDone(t, db, holder, done)
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !finished {
		result = <-done
	}

	if got := countReplays(t, db, firstID) + countReplays(t, db, secondID); got != 1 {
		t.Errorf("replays of the forked siblings after two concurrent replays = %d, want 1", got)
	}
	wantRefused(t, result, secondID, fmt.Sprintf("job lineage cannot be replayed: job %d is pending", competingID))
}

func TestReplayCommandReplaysEitherFailedSiblingOfAForkedLineage(t *testing.T) {
	for _, replayed := range []string{"first", "second"} {
		t.Run(replayed, func(t *testing.T) {
			db := workerTestDatabase(t)
			originalID := insertLineageJob(t, db, "failed", 0)
			siblings := map[string]int64{
				"first":  insertLineageJob(t, db, "failed", originalID),
				"second": insertLineageJob(t, db, "failed", originalID),
			}

			err := replayJob(t, siblings[replayed])

			if err != nil {
				t.Fatalf("replay of the %s failed sibling: %v", replayed, err)
			}
			if got := countReplays(t, db, siblings[replayed]); got != 1 {
				t.Errorf("replays of the %s failed sibling = %d, want 1", replayed, got)
			}
		})
	}
}
