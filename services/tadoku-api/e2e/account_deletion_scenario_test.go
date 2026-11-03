package e2e_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

const deletionRole = "deletion"

// newDeletionScenario provisions one deletion tenant holding the account cast, with Admin One as administrator.
func newDeletionScenario(t *testing.T) *scenario {
	t.Helper()
	s := newScenario(t, deletionRole)
	s.seed(t, deletionRole, tenantSeed{relationships: []string{"testdata/ScenarioIsolation/relationships.json"}})
	return s
}

func (s *scenario) fileReaderDeletion(t *testing.T) {
	t.Helper()
	response := s.serve(t, deletionRole, admin, readerDeletionRequest())
	if response.Code != http.StatusAccepted {
		t.Fatalf("filing: status=%d body=%s, want 202", response.Code, response.Body)
	}
}

// enqueueReaderDeletionJob inserts account.delete.v1 for the reader as fixture SQL, so it reaches the worker without a
// filing.
func (s *scenario) enqueueReaderDeletionJob(t *testing.T) {
	t.Helper()
	_, err := api.db.Pool.Exec(t.Context(), `insert into jobs (tenant, task_type, payload, created_at, next_attempt_at)
		values ($1, 'account.delete.v1', $2, $3, $3)`,
		s.tenant(t, deletionRole).String(),
		`{"user_id":"`+readerUserID+`"}`,
		s.clock.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func (s *scenario) runJobs(t *testing.T) (failed int) {
	t.Helper()
	stop := s.startWorker(t, deletionRole)
	defer stop()
	return waitForWorker(t, api, s.tenant(t, deletionRole).String())
}

// checkVerifyGolden runs directory's verify.sql bound to the deletion tenant, so row-level security confines every
// table it reads.
func (s *scenario) checkVerifyGolden(t *testing.T, directory string) {
	t.Helper()
	query, err := os.ReadFile(filepath.Join(directory, "verify.sql"))
	if err != nil {
		t.Fatalf("read verify query: %v", err)
	}

	tx, err := api.db.AppPool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	_, err = tx.Exec(t.Context(), "select set_config('tadoku.tenant', $1, true)", s.tenant(t, deletionRole).String())
	if err != nil {
		t.Fatal(err)
	}
	var rows string
	if err := tx.QueryRow(t.Context(), wrapVerifyQuery(string(query))).Scan(&rows); err != nil {
		t.Fatalf("run verify query: %v", err)
	}
	checkJSONGolden(t, directory, []byte(rows))
}
