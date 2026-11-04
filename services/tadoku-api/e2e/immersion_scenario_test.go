package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// normalizeLogUnitIDs gives every migration-seeded log unit the id md5(unit_key || ':' || language_code), once per
// test database, so fixtures and goldens name units by stable ids without writing the shared log_units table.
func normalizeLogUnitIDs(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `update log_units
		set id = md5(unit_key || ':' || coalesce(language_code, ''))::uuid
		where id <> md5(unit_key || ':' || coalesce(language_code, ''))::uuid`)
	if err != nil {
		return fmt.Errorf("normalize log unit ids: %w", err)
	}
	return nil
}

// resetAfter resets the database when the serial top-level test t ends. Call it from a serial test that writes shared
// reference data, such as languages, so the parallel scenarios that start after every serial test read the baseline.
func resetAfter(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := api.db.Reset(ctx); err != nil {
			t.Error(err)
		}
	})
}
