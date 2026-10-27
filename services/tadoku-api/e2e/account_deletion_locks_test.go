package e2e_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestAccountDeletionFilingWaitsForAccountWrites(t *testing.T) {
	tokens := resetAccountDeletion(t)
	holder := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)

	atFixtureInstant(func() {
		filed := serveAsync(readerDeletion(tokens))
		waitForLockWaiters(t, 1, filed)

		if err := holder.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireBody(t, <-filed, http.StatusAccepted, "")
	})
}

func TestAccountDeletionConfirmsTheLockedAddress(t *testing.T) {
	tokens := resetAccountDeletion(t)
	change := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	_, err := change.Exec(t.Context(), `update accounts set email = 'renamed@example.test',
		email_normalized = 'renamed@example.test' where user_id = $1`, readerUserID)
	if err != nil {
		t.Fatal(err)
	}

	atFixtureInstant(func() {
		filed := serveAsync(readerDeletion(tokens))
		waitForLockWaiters(t, 1, filed)

		if err := change.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireBody(t, <-filed, http.StatusBadRequest, `"error":"confirmation_mismatch"`)
	})

	checkVerifyGolden(t, api, refusedFilingDir)
}

func TestAccountDeletionQueuedSettingsWriteFails(t *testing.T) {
	tokens := resetAccountDeletion(t)
	login := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login", readerLogin, "", nil)
	requireIdentityStatus(t, login, http.StatusOK)
	rename := newIdentityRequest(t, "tadoku/prod", "PATCH", "/identity/settings/profile",
		`{"display_name":"Renamed Reader"}`, login.Result().Cookies()[0].String(), nil)
	holder := holdRow(t, "select 1 from users where id = $1 for update", readerUserID)

	atFixtureInstant(func() {
		filed := serveAsync(readerDeletion(tokens))
		waitForLockWaiters(t, 1, filed)
		renamed := serveAsync(rename)
		waitForLockWaiters(t, 2, filed)

		if err := holder.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireBody(t, <-filed, http.StatusAccepted, "")
		if response := <-renamed; response.Code != http.StatusUnauthorized {
			t.Errorf("settings write queued behind filing: status=%d body=%s, want 401", response.Code, response.Body)
		}
	})

	checkVerifyGolden(t, api, "testdata/AccountDeletion/settings_write_refused")
}

func TestAccountDeletionRefilingDoesNotWaitForTheWorker(t *testing.T) {
	tokens := resetAccountDeletion(t)
	fileReaderDeletion(t, tokens)
	deletion := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	if _, err := deletion.Exec(t.Context(), "select 1 from users where id = $1 for update", readerUserID); err != nil {
		t.Fatal(err)
	}

	atFixtureInstant(func() {
		refiled := serveAsync(readerDeletion(tokens))

		select {
		case response := <-refiled:
			requireBody(t, response, http.StatusConflict, `"error":"account_deletion_in_progress"`)
		case <-time.After(5 * time.Second):
			t.Error("filing again waits for the running deletion")
			_ = deletion.Rollback(t.Context())
			<-refiled
		}
	})
}

func TestAccountDeletionFilingQueuedBehindFilingTakesNoUserLock(t *testing.T) {
	tokens := resetAccountDeletion(t)
	first := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	_, err := first.Exec(t.Context(), `update users set deletion_locked_at = '2026-09-12 12:00:00'
		where id = $1`, readerUserID)
	if err != nil {
		t.Fatal(err)
	}

	atFixtureInstant(func() {
		filed := serveAsync(readerDeletion(tokens))
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
		case response := <-filed:
			requireBody(t, response, http.StatusConflict, `"error":"account_deletion_in_progress"`)
		case <-time.After(5 * time.Second):
			t.Error("the second filing waits for the users row another transaction holds")
			_ = holder.Rollback(t.Context())
			<-filed
		}
	})
}

func TestAccountDeletionWorkerLocksTheAccountBeforeTheUser(t *testing.T) {
	tokens := resetAccountDeletion(t)
	fileReaderDeletion(t, tokens)
	login := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)

	atFixtureInstant(func() {
		stop := startWorker(t, api, "")
		defer stop()
		waitForLockWaiters(t, 1, nil)

		probe := holdRow(t, "select 1")
		_, err := probe.Exec(t.Context(), "select 1 from users where id = $1 for update nowait", readerUserID)
		if err != nil {
			t.Errorf("the worker holds the users row while it waits for the account row: %v", err)
		}
		if err := probe.Rollback(t.Context()); err != nil {
			t.Fatal(err)
		}

		if err := login.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		if failed := waitForWorker(t, api, ""); failed != 0 {
			t.Errorf("worker failed %d jobs after the account row was released", failed)
		}
	})
}

func TestAccountDeletionRoleChangeWaitsForTheWorker(t *testing.T) {
	tokens := resetAccountDeletion(t)
	fileReaderDeletion(t, tokens)
	deletion := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	for _, statement := range []string{
		"update users set display_name = 'Deleted participant', deleted_at = '2026-09-12 12:00:00' where id = $1",
		"delete from permission_relationships where subject_type = 'user' and subject_id = $1::text",
	} {
		if _, err := deletion.Exec(t.Context(), statement, readerUserID); err != nil {
			t.Fatal(err)
		}
	}
	ban := httptest.NewRequest(http.MethodPut, "/authz/users/"+readerUserID+"/role",
		strings.NewReader(`{"role":"banned","reason":"repeated abuse"}`))
	ban.Header.Set("Content-Type", "application/json")
	authorize(ban, admin, tokens)

	atFixtureInstant(func() {
		roleChanged := serveAsync(ban)
		waitForLockWaiters(t, 1, roleChanged)

		if err := deletion.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireBody(t, <-roleChanged, http.StatusConflict, `"error":"account_deletion_in_progress"`)
	})

	checkVerifyGolden(t, api, "testdata/AccountDeletion/role_change_refused")
}

func TestAccountDeletionFilingSeesARoleChangeHoldingTheUser(t *testing.T) {
	tokens := resetAccountDeletion(t)
	change := holdRow(t, "select 1 from users where id = $1 for update", readerUserID)
	_, err := change.Exec(t.Context(), `insert into permission_relationships
		(tenant, object_type, object_id, relation, subject_type, subject_id)
		values ('tadoku/prod', 'app', 'root', 'admins', 'user', $1)`, readerUserID)
	if err != nil {
		t.Fatal(err)
	}

	atFixtureInstant(func() {
		filed := serveAsync(readerDeletion(tokens))
		waitForLockWaiters(t, 1, filed)

		if err := change.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireBody(t, <-filed, http.StatusConflict, `"error":"admin_account"`)
	})

	checkVerifyGolden(t, api, refusedFilingDir)
}

// serveAsync serves request through api.handler on another goroutine. Call it inside atFixtureInstant and receive the
// response before that scope ends, so that the request never runs while the clocks are restored.
func serveAsync(request *http.Request) <-chan *httptest.ResponseRecorder {
	served := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		response := httptest.NewRecorder()
		api.handler.ServeHTTP(response, request)
		served <- response
	}()
	return served
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

// waitForLockWaiters fails the test if the request behind served finishes before n backends wait on a row lock.
func waitForLockWaiters(t *testing.T, n int, served <-chan *httptest.ResponseRecorder) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		select {
		case response := <-served:
			t.Fatalf("finished (status=%d body=%s) while another transaction held its row", response.Code, response.Body)
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
