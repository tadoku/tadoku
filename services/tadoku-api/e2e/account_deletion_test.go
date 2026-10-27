package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	accountDeletionJourneyDir = "testdata/journeys/AdminAccountDeletion"
	refusedFilingDir          = "testdata/AccountDeletion/refused_filing"
	readerUserID              = "11111111-1111-4111-8111-111111111111"
	readerLogin               = `{"email":"reader@example.test","password":"` + identityFixturePassword + `"}`
)

func TestAccountDeletionJobRequiresLock(t *testing.T) {
	resetAccountDeletion(t)
	enqueueReaderDeletionJob(t)

	if failed := runJobs(t); failed != 1 {
		t.Errorf("worker failed %d jobs, want only the job for a user without a deletion lock", failed)
	}
	checkVerifyGolden(t, api, "testdata/AccountDeletion/unlocked_job")
}

func TestAccountDeletionReplayedJobChangesNothing(t *testing.T) {
	tokens := resetAccountDeletion(t)
	fileReaderDeletion(t, tokens)
	if failed := runJobs(t); failed != 0 {
		t.Fatalf("worker failed %d jobs of the deletion", failed)
	}
	checkVerifyGolden(t, api, "testdata/AccountDeletion/reader_deleted")

	enqueueReaderDeletionJob(t)

	if failed := runJobs(t); failed != 0 {
		t.Errorf("worker failed %d jobs, want the replayed deletion to complete", failed)
	}
	checkVerifyGolden(t, api, "testdata/AccountDeletion/reader_deletion_replayed")
}

func TestAccountDeletionFreesTheAddress(t *testing.T) {
	tokens := resetAccountDeletion(t)
	fileReaderDeletion(t, tokens)
	if failed := runJobs(t); failed != 0 {
		t.Fatalf("worker failed %d jobs of the deletion", failed)
	}

	response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/registration",
		`{"email":"reader@example.test","display_name":"New Reader","password":"Independent-653!"}`, "", nil)

	requireIdentityStatus(t, response, http.StatusCreated)
	checkVerifyGolden(t, api, "testdata/AccountDeletion/address_registered_again")
}

// enqueueReaderDeletionJob inserts account.delete.v1 for the reader as fixture SQL, so it reaches the worker without a
// filing.
func enqueueReaderDeletionJob(t *testing.T) {
	t.Helper()
	_, err := api.db.Pool.Exec(t.Context(), `insert into jobs (tenant, task_type, payload, created_at, next_attempt_at)
		values ('tadoku/prod', 'account.delete.v1', $1, $2, $2)`,
		`{"user_id":"`+readerUserID+`"}`,
		fixtureInstant,
	)
	if err != nil {
		t.Fatal(err)
	}
}

// runJobs runs the journey worker at the fixture instant and returns the number of jobs that failed.
func runJobs(t *testing.T) (failed int) {
	t.Helper()
	atFixtureInstant(func() {
		stop := startWorker(t, api, "")
		defer stop()
		failed = waitForWorker(t, api, "")
	})
	return failed
}

func resetAccountDeletion(t *testing.T) map[member]string {
	t.Helper()
	resetJourney(t, api, accountDeletionJourneyDir)
	tokens, err := loadCast(nil)
	if err != nil {
		t.Fatal(err)
	}
	return tokens
}

// readerDeletion builds the admin's filing for the reader, confirmed with the address testdata/accounts.sql gives it.
func readerDeletion(tokens map[member]string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/profile/users/"+readerUserID+"/account-deletion",
		strings.NewReader(`{"email":"reader@example.test","reason":"Owner confirmed by email"}`))
	request.Header.Set("Content-Type", "application/json")
	authorize(request, admin, tokens)
	return request
}

func fileReaderDeletion(t *testing.T, tokens map[member]string) {
	t.Helper()
	response := httptest.NewRecorder()
	atFixtureInstant(func() { api.handler.ServeHTTP(response, readerDeletion(tokens)) })
	if response.Code != http.StatusAccepted {
		t.Fatalf("filing: status=%d body=%s, want 202", response.Code, response.Body)
	}
}

func requireBody(t *testing.T, response *httptest.ResponseRecorder, status int, fragment string) {
	t.Helper()
	if response.Code != status || !strings.Contains(response.Body.String(), fragment) {
		t.Errorf("status=%d body=%s, want %d with %s", response.Code, response.Body, status, fragment)
	}
}
