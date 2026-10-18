package jobqueue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

var jobTestTime = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

func jobTestDB(t *testing.T) *testpostgres.Database {
	t.Helper()
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db, err := testpostgres.New(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

func at(t *testing.T, instant time.Time, work func()) {
	t.Helper()
	timex.TheWorld(instant, work)
}

func insertJob(repo *Repository, ctx context.Context, job jobs.Job) (int64, error) {
	payload, err := json.Marshal(job)
	if err != nil {
		return 0, err
	}
	return repo.Insert(ctx, job.Type(), payload)
}
