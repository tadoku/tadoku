package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	queries "github.com/tadoku/tadoku/services/tadoku-api/infra/postgres/internal/pgxcompat"
)

// Real sqlc output accepts the same native pool/transaction surface directly.
var (
	_ postgres.DBTX = (*pgxpool.Pool)(nil)
	_ postgres.DBTX = (pgx.Tx)(nil)
	_ queries.DBTX  = (postgres.DBTX)(nil)
	_ postgres.DBTX = (queries.DBTX)(nil)
)

func TestSQLCUsesNativeExecutorWithoutAdapter(t *testing.T) {
	t.Parallel()
	pool := openPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	refused := errors.New("synthetic rollback")
	for _, rollback := range []bool{false, true} {
		// A temporary table is scoped to the transaction's connection; no
		// permanent application table or migration is created by this test.
		err := postgres.RunInTransaction(ctx, pool, func(ctx context.Context) error {
			db, err := postgres.Executor(ctx, pool)
			if err != nil {
				return err
			}
			if _, err := db.Exec(ctx, "create temporary table query_compat_items (id integer primary key, label text not null) on commit drop"); err != nil {
				return err
			}
			q := queries.New(db) // No cast, adapter or generated-source edit.
			if err := q.InsertItem(ctx, queries.InsertItemParams{ID: 1, Label: "before"}); err != nil {
				return err
			}
			if err := q.RenameItem(ctx, queries.RenameItemParams{ID: 1, Label: "after"}); err != nil {
				return err
			}
			item, err := q.GetItem(ctx, 1)
			if err != nil {
				return err
			}
			if item.Label != "after" {
				t.Errorf("composed generated queries returned label=%q, want after", item.Label)
			}
			items, err := q.ListItems(ctx)
			if err != nil {
				return err
			}
			if len(items) != 1 || items[0] != item {
				t.Errorf("generated multi-row query returned %v, want [%v]", items, item)
			}
			if rollback {
				return refused
			}
			return nil
		})
		if rollback && !errors.Is(err, refused) {
			t.Errorf("generated-query rollback error=%v, want synthetic refusal", err)
		}
		if !rollback && err != nil {
			t.Errorf("generated-query commit: %v", err)
		}
	}
	// Pool-based execution also passes directly to generated bindings.
	db, err := postgres.Executor(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	_ = queries.New(db)
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("pool reuse after generated-query scopes: %v", err)
	}
}
