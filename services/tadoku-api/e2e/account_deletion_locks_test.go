package e2e_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	featureprofile "github.com/tadoku/tadoku/services/tadoku-api/features/profile"
)

func TestAccountDeletionFilingWaitsForAccountWrites(t *testing.T) {
	resetAccountDeletion(t)
	holder := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	file := operatorFiling(t, readerUserID, adminUserID, "Owner confirmed by email")

	filed := make(chan error, 1)
	go func() {
		_, err := file()
		filed <- err
	}()
	waitForLockWaiters(t, 1, filed)

	if err := holder.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := <-filed; err != nil {
		t.Fatal(err)
	}
}

func TestAccountDeletionQueuedSettingsWriteFails(t *testing.T) {
	resetAccountDeletion(t)
	cookie := productionLoginCookie(t, "reader@example.test")
	holder := holdRow(t, "select 1 from users where id = $1 for update", readerUserID)
	file := operatorFiling(t, readerUserID, adminUserID, "Owner confirmed by email")
	rename := newIdentityRequest(t, "tadoku/prod", "PATCH", "/identity/settings/profile",
		`{"display_name":"Renamed Reader"}`, cookie, nil)

	filed := make(chan error, 1)
	go func() {
		_, err := file()
		filed <- err
	}()
	waitForLockWaiters(t, 1, filed)
	renamed := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		response := httptest.NewRecorder()
		atFixtureInstant(func() { api.handler.ServeHTTP(response, rename) })
		renamed <- response
	}()
	waitForLockWaiters(t, 2, filed)

	if err := holder.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := <-filed; err != nil {
		t.Fatal(err)
	}
	response := <-renamed
	if response.Code != http.StatusUnauthorized {
		t.Errorf("settings write queued behind filing: status=%d body=%s, want 401", response.Code, response.Body)
	}
	if name := snapshot(t, "select display_name from users where tenant = 'tadoku/prod' and id = $1", readerUserID); name != "Reader One" {
		t.Errorf("display name=%q after a settings write queued behind filing", name)
	}
}

func TestAccountDeletionRefilingDoesNotWaitForTheWorker(t *testing.T) {
	resetAccountDeletion(t)
	if _, err := fileOperatorDeletion(t, readerUserID, adminUserID, "Owner confirmed by email"); err != nil {
		t.Fatal(err)
	}
	deletion := holdRow(t, "select 1 from users where id = $1 for update", readerUserID)
	file := operatorFiling(t, readerUserID, adminUserID, "Asked again")

	filed := make(chan error, 1)
	go func() {
		_, err := file()
		filed <- err
	}()
	select {
	case err := <-filed:
		if !errors.Is(err, featureprofile.ErrAccountDeletionInProgress) {
			t.Errorf("filing again returned %v, want %v", err, featureprofile.ErrAccountDeletionInProgress)
		}
	case <-time.After(5 * time.Second):
		t.Error("filing again waits for the users row the worker holds")
	}

	if err := deletion.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestAccountDeletionFilingQueuedBehindFilingTakesNoUserLock(t *testing.T) {
	resetAccountDeletion(t)
	first := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	if _, err := first.Exec(t.Context(), `update users set deletion_locked_at = '2026-09-12 12:00:00'
		where id = $1`, readerUserID); err != nil {
		t.Fatal(err)
	}
	file := operatorFiling(t, readerUserID, adminUserID, "Asked twice")

	filed := make(chan error, 1)
	go func() {
		_, err := file()
		filed <- err
	}()
	waitForLockWaiters(t, 1, filed)
	deletion := holdRowAsync(t, "select 1 from users where id = $1 for update", readerUserID)
	waitForLockWaiters(t, 2, filed)

	if err := first.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	deletionTx, held := <-deletion
	if !held {
		t.Fatal("the stand-in worker could not lock the users row")
	}
	select {
	case err := <-filed:
		if !errors.Is(err, featureprofile.ErrAccountDeletionInProgress) {
			t.Errorf("the second filing returned %v, want %v", err, featureprofile.ErrAccountDeletionInProgress)
		}
	case <-time.After(5 * time.Second):
		t.Error("the second filing waits for the users row the worker holds")
	}
	if err := deletionTx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
}

// holdRowAsync starts holdRow's transaction on another goroutine, for a lock that has to wait, and delivers it once
// the lock is held. It fails the test from the test goroutine only.
func holdRowAsync(t *testing.T, query string, arguments ...any) <-chan pgx.Tx {
	t.Helper()
	tx, err := api.db.AppPool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	if _, err := tx.Exec(t.Context(), "select set_config('tadoku.tenant', 'tadoku/prod', true)"); err != nil {
		t.Fatal(err)
	}

	held := make(chan pgx.Tx, 1)
	go func() {
		if _, err := tx.Exec(context.Background(), query, arguments...); err == nil {
			held <- tx
		}
		close(held)
	}()
	return held
}

// holdRow runs the locking query in a tadoku/prod transaction that stays open until the caller commits it or the
// test ends.
func holdRow(t *testing.T, query string, arguments ...any) pgx.Tx {
	t.Helper()
	tx, err := api.db.AppPool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })

	if _, err := tx.Exec(t.Context(), "select set_config('tadoku.tenant', 'tadoku/prod', true)"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(t.Context(), query, arguments...); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return tx
}

// waitForLockWaiters fails the test if filing finishes before n backends wait on a row lock.
func waitForLockWaiters(t *testing.T, n int, filed <-chan error) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		select {
		case err := <-filed:
			t.Fatalf("filing finished (%v) while another transaction held its row", err)
		default:
		}

		var waiting int
		err := api.db.Pool.QueryRow(t.Context(), `select count(*) from pg_stat_activity
			where datname = current_database() and wait_event_type = 'Lock'`).Scan(&waiting)
		if err != nil {
			t.Fatal(err)
		}
		if waiting >= n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d backends wait on a row lock, want %d", waiting, n)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
