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

func TestAccountDeletionRefilingTakesNoAccountLock(t *testing.T) {
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
	waitForLockWaiters(t, 1, filed)
	probe := holdRow(t, "select 1")
	_, err := probe.Exec(t.Context(), "select 1 from accounts where user_id = $1 for update nowait", readerUserID)
	if err != nil {
		t.Errorf("filing again holds the account row while it waits for the users row: %v", err)
	}
	if err := probe.Rollback(t.Context()); err != nil {
		t.Fatal(err)
	}

	if err := deletion.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := <-filed; !errors.Is(err, featureprofile.ErrAccountDeletionInProgress) {
		t.Errorf("filing again returned %v, want %v", err, featureprofile.ErrAccountDeletionInProgress)
	}
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
