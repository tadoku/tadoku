package jobqueue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
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

func replayJob(ctx context.Context, pool *pgxpool.Pool, id int64, supported ...jobs.Type) (int64, error) {
	var replayID int64
	err := postgres.RunInTransaction(ctx, pool, func(ctx context.Context) error {
		var err error
		replayID, err = NewService(NewRepository(pool)).Replay(ctx, id, "operator", "repair", supported)
		return err
	})
	return replayID, err
}
