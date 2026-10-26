package e2e_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	featureprofile "github.com/tadoku/tadoku/services/tadoku-api/features/profile"
)

func TestAccountDeletionRefusesAdministrator(t *testing.T) {
	resetAccountDeletion(t)
	before := snapshot(t, targetSnapshotSQL, adminUserID)

	_, err := fileAdminDeletion(t, adminUserID, bystanderUserID, "Owner confirmed by email")

	if !errors.Is(err, identity.ErrAdminAccount) {
		t.Fatalf("filing for an administrator returned %v, want %v", err, identity.ErrAdminAccount)
	}
	if after := snapshot(t, targetSnapshotSQL, adminUserID); after != before {
		t.Errorf("refused filing changed state\nbefore=%s\nafter=%s", before, after)
	}
	if jobs := countJobs(t); jobs != 0 {
		t.Errorf("refused filing enqueued %d jobs", jobs)
	}
}

func TestAccountDeletionFilingEndsAccess(t *testing.T) {
	resetAccountDeletion(t)
	cookie := productionLoginCookie(t, "reader@example.test")
	recoveryToken := seedProductionRecoveryToken(t, readerUserID)

	jobID, err := fileAdminDeletion(t, readerUserID, adminUserID, "Owner confirmed by email")

	if err != nil {
		t.Fatal(err)
	}
	if jobID < 1 {
		t.Errorf("job ID=%d, want the enqueued account.delete.v1 ID", jobID)
	}
	checkVerifyGolden(t, api, accountDeletionDir+"/filed")

	response := identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/session", "", cookie, nil)
	if identityView(t, response).Account != nil {
		t.Error("the session of a user with a pending deletion still authenticates")
	}
	response = identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login",
		`{"email":"reader@example.test","password":"`+identityFixturePassword+`"}`, "", nil)
	requireErrorBody(t, response, http.StatusUnauthorized, `"code":"wrong_password"`)
	response = identityRequest(t, api.handler, "tadoku/prod", "GET", "/identity/recovery/"+recoveryToken, "", "", nil)
	requireErrorBody(t, response, http.StatusBadRequest, `"code":"invalid_recovery_token"`)

	tokens, err := loadCast(nil)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("POST", "/immersion/logs", strings.NewReader(
		`{"language_code":"eng","activity_id":1,"duration_seconds":60,"tags":[]}`,
	))
	request.Header.Set("Content-Type", "application/json")
	authorize(request, user, tokens)
	response = httptest.NewRecorder()
	atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })
	requireErrorBody(t, response, http.StatusConflict, `"error":"account_deletion_in_progress"`)
}

func TestAccountDeletionRemovesData(t *testing.T) {
	resetAccountDeletion(t)
	productionLoginCookie(t, "banned@example.test")
	productionLoginCookie(t, "user2@example.test")
	seedProductionRecoveryToken(t, bannedUserID)
	others := snapshot(t, othersSnapshotSQL, bannedUserID)
	history := snapshot(t, finishedHistorySQL)

	_, err := fileAdminDeletion(t, bannedUserID, adminUserID, "Owner confirmed by email")
	if err != nil {
		t.Fatal(err)
	}
	runDeletionWorker(t)

	checkVerifyGolden(t, api, accountDeletionDir+"/deleted")
	checkVerifyGolden(t, api, accountDeletionDir+"/jobs")
	checkVerifyGolden(t, api, accountDeletionDir+"/contest_events")
	checkVerifyGolden(t, api, accountDeletionDir+"/audit")
	checkVerifyGolden(t, api, accountDeletionDir+"/emails")
	if after := snapshot(t, othersSnapshotSQL, bannedUserID); after != others {
		t.Errorf("deletion changed rows of other users or tenants\nbefore=%s\nafter=%s", others, after)
	}
	if after := snapshot(t, finishedHistorySQL); after != history {
		t.Errorf("deletion changed finished-contest scores\nbefore=%s\nafter=%s", history, after)
	}

	deleted := snapshot(t, targetSnapshotSQL, bannedUserID)
	jobs := countJobs(t)
	enqueueAccountDeletionJob(t, bannedUserID)
	runDeletionWorker(t)
	if after := snapshot(t, targetSnapshotSQL, bannedUserID); after != deleted {
		t.Errorf("a duplicate job changed the deleted user\nbefore=%s\nafter=%s", deleted, after)
	}
	if after := countJobs(t); after != jobs+1 {
		t.Errorf("a duplicate job left %d jobs, want %d", after, jobs+1)
	}
	checkVerifyGolden(t, api, accountDeletionDir+"/audit")

	_, err = fileAdminDeletion(t, bannedUserID, adminUserID, "Asked again")
	if !errors.Is(err, featureprofile.ErrLocalUserNotFound) {
		t.Errorf("filing for the deleted user returned %v, want %v", err, featureprofile.ErrLocalUserNotFound)
	}
	if after := snapshot(t, targetSnapshotSQL, bannedUserID); after != deleted {
		t.Errorf("filing for the deleted user changed it\nbefore=%s\nafter=%s", deleted, after)
	}
	if after := countJobs(t); after != jobs+1 {
		t.Errorf("filing for the deleted user left %d jobs, want %d", after, jobs+1)
	}

	response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/registration",
		`{"email":"banned@example.test","display_name":"New Reader","password":"Independent-653!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusCreated)
	var registered string
	err = api.db.Pool.QueryRow(t.Context(), `select user_id::text from accounts
		where tenant = 'tadoku/prod' and email_normalized = 'banned@example.test'`).Scan(&registered)
	if err != nil {
		t.Fatal(err)
	}
	if registered == bannedUserID {
		t.Error("registering the freed address reused the deleted user ID")
	}
}

func TestAccountDeletionJobRequiresLock(t *testing.T) {
	resetAccountDeletion(t)
	before := snapshot(t, targetSnapshotSQL, bystanderUserID)
	enqueueAccountDeletionJob(t, bystanderUserID)

	runDeletionWorker(t)

	var state, lastError string
	err := api.db.Pool.QueryRow(t.Context(), `select state, coalesce(last_error, '') from jobs
		where task_type = 'account.delete.v1'`).Scan(&state, &lastError)
	if err != nil {
		t.Fatal(err)
	}
	if state != "failed" {
		t.Errorf("job for an unlocked user is %s (%s), want failed", state, lastError)
	}
	if after := snapshot(t, targetSnapshotSQL, bystanderUserID); after != before {
		t.Errorf("job for an unlocked user changed state\nbefore=%s\nafter=%s", before, after)
	}
	if jobs := countJobs(t); jobs != 1 {
		t.Errorf("job for an unlocked user left %d jobs, want only itself", jobs)
	}
}

func countJobs(t *testing.T) int {
	t.Helper()
	var count int
	if err := api.db.Pool.QueryRow(t.Context(), "select count(*) from jobs").Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func requireErrorBody(t *testing.T, response *httptest.ResponseRecorder, status int, fragment string) {
	t.Helper()
	if response.Code != status || !strings.Contains(response.Body.String(), fragment) {
		t.Errorf("status=%d body=%s, want %d with %s", response.Code, response.Body, status, fragment)
	}
}
