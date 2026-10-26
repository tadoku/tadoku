package e2e_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
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
	deletion := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	if _, err := deletion.Exec(t.Context(), "select 1 from users where id = $1 for update", readerUserID); err != nil {
		t.Fatal(err)
	}
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
		t.Error("filing again waits for the running deletion")
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
	usersHolder := holdRowAsync(t, "select 1 from users where id = $1 for update", readerUserID)
	waitForLockWaiters(t, 2, filed)

	if err := first.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	holder, held := <-usersHolder
	if !held {
		t.Fatal("the second transaction could not lock the users row")
	}
	select {
	case err := <-filed:
		if !errors.Is(err, featureprofile.ErrAccountDeletionInProgress) {
			t.Errorf("the second filing returned %v, want %v", err, featureprofile.ErrAccountDeletionInProgress)
		}
	case <-time.After(5 * time.Second):
		t.Error("the second filing waits for the users row another transaction holds")
	}
	if err := holder.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestAccountDeletionWorkerLocksTheAccountBeforeTheUser(t *testing.T) {
	resetAccountDeletion(t)
	if _, err := fileOperatorDeletion(t, readerUserID, adminUserID, "Owner confirmed by email"); err != nil {
		t.Fatal(err)
	}
	login := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)

	scope, err := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
	if err != nil {
		t.Fatal(err)
	}
	application := newWorkerApplication(t, api.db.AppPool, scope)
	ctx, stop := context.WithCancel(t.Context())
	stopped := make(chan error, 1)
	go atFixtureInstant(func() { stopped <- application.Run(ctx) })
	defer func() {
		stop()
		<-stopped
	}()
	waitForLockWaiters(t, 1, nil)

	probe := holdRow(t, "select 1")
	if _, err := probe.Exec(t.Context(), "select 1 from users where id = $1 for update nowait", readerUserID); err != nil {
		t.Errorf("the worker holds the users row while it waits for the account row: %v", err)
	}
	if err := probe.Rollback(t.Context()); err != nil {
		t.Fatal(err)
	}

	if err := login.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for state := ""; state != "completed"; time.Sleep(20 * time.Millisecond) {
		err := api.db.Pool.QueryRow(t.Context(), `select state from jobs
			where task_type = 'account.delete.v1'`).Scan(&state)
		if err != nil {
			t.Fatal(err)
		}
		if state == "failed" || time.Now().After(deadline) {
			t.Fatalf("account.delete.v1 is %s after the account row was released", state)
		}
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
