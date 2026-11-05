package leaderboard

import (
	"context"
	"errors"
	"maps"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

type versionsWriter struct {
	pid      uint32
	finished chan error
	release  chan struct{}
}

func beginVersionsTx(t *testing.T, db *testpostgres.Database, key string) pgx.Tx {
	t.Helper()

	tx, err := db.AppPool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.WithoutCancel(t.Context())) })
	if _, err := tx.Exec(t.Context(), "select set_config('tadoku.tenant', $1, true)", key); err != nil {
		t.Fatal(err)
	}
	return tx
}

// startVersionsWriter runs sql in its own transaction and commits once release is closed.
func startVersionsWriter(t *testing.T, db *testpostgres.Database, key, sql string, args ...any) *versionsWriter {
	t.Helper()

	tx := beginVersionsTx(t, db, key)
	writer := &versionsWriter{
		pid:      tx.Conn().PgConn().PID(),
		finished: make(chan error, 1),
		release:  make(chan struct{}),
	}
	go func() {
		if _, err := tx.Exec(t.Context(), sql, args...); err != nil {
			writer.finished <- err
			return
		}
		<-writer.release
		writer.finished <- tx.Commit(t.Context())
	}()
	return writer
}

// waitUntilSettled returns once the backend's current or last statement is sql and it waits on a lock or has
// finished.
func waitUntilSettled(t *testing.T, db *testpostgres.Database, pid uint32, sql string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var settled bool
		err := db.Pool.QueryRow(t.Context(), `select coalesce(bool_or(wait_event_type = 'Lock'
			or state like 'idle in transaction%'), false)
			from pg_stat_activity
			where pid = $1 and query = $2`,
			pid,
			sql,
		).Scan(&settled)
		if err != nil {
			t.Fatal(err)
		}
		if settled {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("backend %d neither finished %q nor waited on a lock", pid, sql)
}

func TestLeaderboardVersionsLetConcurrentWritersCommit(t *testing.T) {
	t.Parallel()
	db := newVersionsDatabase(t)
	key := tenant.Production().String()
	writeAs(t, db.AppPool, key, insertVersionsUserSQL, versionsReader)
	const writers = 8

	var wg sync.WaitGroup
	errs := make(chan error, writers)
	for writer := range writers {
		wg.Go(func() {
			years := []string{"2025-03-01", "2026-03-01"}
			if writer%2 == 1 {
				years[0], years[1] = years[1], years[0]
			}
			errs <- pgx.BeginFunc(t.Context(), db.AppPool, func(tx pgx.Tx) error {
				if _, err := tx.Exec(t.Context(), "select set_config('tadoku.tenant', $1, true)", key); err != nil {
					return err
				}
				for _, createdAt := range years {
					_, err := tx.Exec(t.Context(), insertVersionsLogSQL, uuid.New(), versionsReader, true, createdAt)
					if err != nil {
						return err
					}
				}
				return nil
			})
		})
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("concurrent writer failed: %v", err)
		}
	}
	want := map[string]int64{"global": 2 * writers, "yearly:2025": writers, "yearly:2026": writers}
	if got := boardVersions(t, db, key); !maps.Equal(got, want) {
		t.Errorf("versions after concurrent writers = %v; want %v", got, want)
	}
}

func TestLeaderboardVersionsLetACreatorAttachWhileAnotherWriterChangesTheContest(t *testing.T) {
	t.Parallel()
	db := newVersionsDatabase(t)
	key := tenant.Production().String()
	seedVersionsBaseline(t, db, key)
	creator := beginVersionsTx(t, db, key)
	_, err := creator.Exec(t.Context(), insertVersionsLogSQL, versionsLog5, versionsReader, true, "2026-01-12")
	if err != nil {
		t.Fatal(err)
	}

	const deleteSQL = `update logs set deleted_at = now() where id = $1`
	deleter := startVersionsWriter(t, db, key, deleteSQL, versionsLog1)
	waitUntilSettled(t, db, deleter.pid, deleteSQL)
	_, err = creator.Exec(t.Context(), insertVersionsContestLogSQL, versionsContestC, versionsLog5)
	if err == nil {
		err = creator.Commit(t.Context())
	}
	close(deleter.release)

	if err != nil {
		t.Errorf("creator attaching its log failed: %v", err)
	}
	if err := <-deleter.finished; err != nil {
		t.Errorf("concurrent deleter failed: %v", err)
	}
}

func TestLeaderboardVersionsLetWritersThatShareARowCommit(t *testing.T) {
	t.Parallel()
	db := newVersionsDatabase(t)
	key := tenant.Production().String()
	seedVersionsBaseline(t, db, key)
	creator := beginVersionsTx(t, db, key)
	_, err := creator.Exec(t.Context(), insertVersionsLogSQL, versionsLog5, versionsReader, true, "2026-01-12")
	if err != nil {
		t.Fatal(err)
	}

	const editSQL = `update logs set computed_score = $2 where id = $1`
	editor := startVersionsWriter(t, db, key, editSQL, versionsLog1, 30)
	waitUntilSettled(t, db, editor.pid, editSQL)
	edited := make(chan error, 1)
	go func() {
		_, err := creator.Exec(t.Context(), editSQL, versionsLog1, 40)
		edited <- err
	}()
	waitUntilSettled(t, db, creator.Conn().PgConn().PID(), editSQL)
	close(editor.release)
	err = <-edited
	if err == nil {
		err = creator.Commit(t.Context())
	}

	if err := <-editor.finished; err != nil {
		t.Errorf("first editor failed: %v", err)
	}
	if err != nil {
		t.Errorf("creator editing the same log failed: %v", err)
	}
}

func TestLeaderboardVersionsKeepATestTenantDeletionFromWaitingOnAWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		finish  func(context.Context, pgx.Tx) error
		tenants int
	}{
		{
			name:    "committed deletion",
			key:     "e2e/versions-0000000e",
			finish:  func(ctx context.Context, tx pgx.Tx) error { return tx.Commit(ctx) },
			tenants: 0,
		},
		{
			name:    "rolled back deletion",
			key:     "e2e/versions-0000000f",
			finish:  func(ctx context.Context, tx pgx.Tx) error { return tx.Rollback(ctx) },
			tenants: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			db := newVersionsDatabase(t)
			if _, err := db.Pool.Exec(t.Context(), `insert into tenants (key, kind) values ($1, 'test')`, test.key); err != nil {
				t.Fatal(err)
			}
			seedVersionsBaseline(t, db, test.key)
			before := boardVersions(t, db, test.key)
			writer := beginVersionsTx(t, db, test.key)
			_, err := writer.Exec(t.Context(), `update logs set created_at = '2019-06-01' where id = $1`, versionsLog1)
			if err != nil {
				t.Fatal(err)
			}

			deletion, err := db.Pool.Begin(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = deletion.Rollback(context.WithoutCancel(t.Context())) })
			const deleteTenantSQL = `delete from tenants where key = $1`
			deleted := make(chan error, 1)
			go func() {
				if _, err := deletion.Exec(t.Context(), deleteTenantSQL, test.key); err != nil {
					deleted <- err
					return
				}
				deleted <- test.finish(t.Context(), deletion)
			}()
			waitUntilSettled(t, db, deletion.Conn().PgConn().PID(), deleteTenantSQL)
			commitErr := writer.Commit(t.Context())

			var pgErr *pgconn.PgError
			if !errors.As(commitErr, &pgErr) || pgErr.Code != "55P03" {
				t.Errorf("writer commit during deletion = %v; want lock_not_available", commitErr)
			}
			if err := <-deleted; err != nil {
				t.Errorf("tenant deletion failed: %v", err)
			}
			var tenants, movedLogs int
			err = db.Pool.QueryRow(t.Context(), `select
				(select count(*) from tenants where key = $1),
				(select count(*) from logs where tenant = $1 and year = 2019)`,
				test.key,
			).Scan(&tenants, &movedLogs)
			if err != nil {
				t.Fatal(err)
			}
			if tenants != test.tenants || movedLogs != 0 {
				t.Errorf("tenants = %d, moved logs = %d; want %d and 0", tenants, movedLogs, test.tenants)
			}
			want := before
			if test.tenants == 0 {
				want = map[string]int64{}
			}
			if got := boardVersions(t, db, test.key); !maps.Equal(got, want) {
				t.Errorf("versions = %v; want %v", got, want)
			}
		})
	}
}
