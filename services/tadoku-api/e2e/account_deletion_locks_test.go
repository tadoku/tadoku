package e2e_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	apiapp "github.com/tadoku/tadoku/services/tadoku-api/app/api"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	featureprofile "github.com/tadoku/tadoku/services/tadoku-api/features/profile"
)

func TestAccountDeletionFilingWaitsForAccountWrites(t *testing.T) {
	resetAccountDeletion(t)
	holder := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	file := adminFiling(t, readerUserID, adminUserID, "Owner confirmed by email")

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

func TestAccountDeletionConfirmsTheLockedAddress(t *testing.T) {
	resetAccountDeletion(t)
	before := snapshot(t, targetSnapshotSQL, readerUserID)
	change := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	_, err := change.Exec(t.Context(), `update accounts set email = 'renamed@example.test',
		email_normalized = 'renamed@example.test' where user_id = $1`, readerUserID)
	if err != nil {
		t.Fatal(err)
	}
	file := adminFiling(t, readerUserID, adminUserID, "Owner confirmed by email")

	filed := make(chan error, 1)
	go func() {
		_, err := file()
		filed <- err
	}()
	waitForLockWaiters(t, 1, filed)
	if err := change.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}

	if err := <-filed; !errors.Is(err, identity.ErrConfirmationMismatch) {
		t.Errorf("filing with the earlier address returned %v, want %v", err, identity.ErrConfirmationMismatch)
	}
	if after := snapshot(t, targetSnapshotSQL, readerUserID); after != before {
		t.Errorf("a refused filing changed state\nbefore=%s\nafter=%s", before, after)
	}
	if jobs := countJobs(t); jobs != 0 {
		t.Errorf("a refused filing enqueued %d jobs", jobs)
	}
}

func TestAccountDeletionQueuedSettingsWriteFails(t *testing.T) {
	resetAccountDeletion(t)
	cookie := productionLoginCookie(t, "reader@example.test")
	holder := holdRow(t, "select 1 from users where id = $1 for update", readerUserID)
	file := adminFiling(t, readerUserID, adminUserID, "Owner confirmed by email")
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
	name := snapshot(t, "select display_name from users where tenant = 'tadoku/prod' and id = $1", readerUserID)
	if name != "Reader One" {
		t.Errorf("display name=%q after a settings write queued behind filing", name)
	}
}

func TestAccountDeletionRefilingDoesNotWaitForTheWorker(t *testing.T) {
	resetAccountDeletion(t)
	if _, err := fileAdminDeletion(t, readerUserID, adminUserID, "Owner confirmed by email"); err != nil {
		t.Fatal(err)
	}
	deletion := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	if _, err := deletion.Exec(t.Context(), "select 1 from users where id = $1 for update", readerUserID); err != nil {
		t.Fatal(err)
	}
	file := adminFiling(t, readerUserID, adminUserID, "Asked again")

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
	file := adminFiling(t, readerUserID, adminUserID, "Asked twice")

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
	if _, err := fileAdminDeletion(t, readerUserID, adminUserID, "Owner confirmed by email"); err != nil {
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
	_, err = probe.Exec(t.Context(), "select 1 from users where id = $1 for update nowait", readerUserID)
	if err != nil {
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

func TestAccountDeletionRoleChangeWaitsForTheWorker(t *testing.T) {
	resetAccountDeletion(t)
	if _, err := fileAdminDeletion(t, readerUserID, adminUserID, "Owner confirmed by email"); err != nil {
		t.Fatal(err)
	}
	deletion := holdRow(t, "select 1 from accounts where user_id = $1 for update", readerUserID)
	for _, statement := range []string{
		"update users set display_name = 'Deleted participant', deleted_at = '2026-09-12 12:00:00' where id = $1",
		"delete from permission_relationships where subject_type = 'user' and subject_id = $1::text",
	} {
		if _, err := deletion.Exec(t.Context(), statement, readerUserID); err != nil {
			t.Fatal(err)
		}
	}
	tokens, err := loadCast(nil)
	if err != nil {
		t.Fatal(err)
	}
	ban := httptest.NewRequest("PUT", "/authz/users/"+readerUserID+"/role",
		strings.NewReader(`{"role":"banned","reason":"repeated abuse"}`))
	ban.Header.Set("Content-Type", "application/json")
	authorize(ban, admin, tokens)

	response := httptest.NewRecorder()
	banned := make(chan error, 1)
	go func() {
		atFixtureInstant(func() { api.handler.ServeHTTP(response, ban) })
		banned <- nil
	}()
	waitForLockWaiters(t, 1, banned)

	if err := deletion.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	<-banned
	if response.Code != http.StatusConflict {
		t.Errorf("banning a user during the deletion: status=%d body=%s, want 409", response.Code, response.Body)
	}
	relationships := snapshot(t, `select count(*)::text from permission_relationships
		where tenant = 'tadoku/prod' and subject_type = 'user' and subject_id = $1::text`, readerUserID)
	if relationships != "0" {
		t.Errorf("the deleted user has %s permission relationships", relationships)
	}
}

func TestAccountDeletionFilingSeesARoleChangeHoldingTheUser(t *testing.T) {
	byAdmin, err := apiapp.AdminAccountDeletion(
		uuid.MustParse(readerUserID),
		uuid.MustParse(adminUserID),
		"reader@example.test",
		"Owner confirmed by email",
	)
	if err != nil {
		t.Fatal(err)
	}
	selfService := apiapp.SelfServiceAccountDeletion(uuid.MustParse(readerUserID))
	cases := []struct {
		name     string
		relation string
		request  apiapp.AccountDeletionRequest
		want     error
	}{
		{"ban during a self-service filing", "banned", selfService, identity.ErrBannedAccount},
		{"admin grant during an admin filing", "admins", byAdmin, identity.ErrAdminAccount},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetAccountDeletion(t)
			before := snapshot(t, targetSnapshotSQL, readerUserID)
			change := holdRow(t, "select 1 from users where id = $1 for update", readerUserID)
			_, err := change.Exec(t.Context(), `insert into permission_relationships
				(tenant, object_type, object_id, relation, subject_type, subject_id)
				values ('tadoku/prod', 'app', 'root', $1, 'user', $2)`, tc.relation, readerUserID)
			if err != nil {
				t.Fatal(err)
			}
			file := deletionFiling(t, tc.request)

			filed := make(chan error, 1)
			go func() {
				_, err := file()
				filed <- err
			}()
			waitForLockWaiters(t, 1, filed)
			if err := change.Commit(t.Context()); err != nil {
				t.Fatal(err)
			}

			if err := <-filed; !errors.Is(err, tc.want) {
				t.Errorf("filing after the role change returned %v, want %v", err, tc.want)
			}
			if after := snapshot(t, targetSnapshotSQL, readerUserID); after != before {
				t.Errorf("a refused filing changed state\nbefore=%s\nafter=%s", before, after)
			}
			if jobs := countJobs(t); jobs != 0 {
				t.Errorf("a refused filing enqueued %d jobs", jobs)
			}
		})
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

// waitForLockWaiters fails the test if the operation behind done finishes before n backends wait on a row lock.
func waitForLockWaiters(t *testing.T, n int, done <-chan error) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		select {
		case err := <-done:
			t.Fatalf("finished (%v) while another transaction held its row", err)
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
