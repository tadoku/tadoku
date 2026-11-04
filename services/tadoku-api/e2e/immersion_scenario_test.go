package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// normalizeLogUnitIDs gives every migration-seeded log unit the id md5(key || ':' || language_code), once per test
// database, so fixtures and goldens name units by stable ids without writing the shared log_units table. The key is
// the activity-prefixed spelling, such as reading_page, whichever spelling the row carries.
func normalizeLogUnitIDs(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `with stable as (
		select
			id,
			md5(
				case
					when unit_key in ('pages', 'sentences', 'characters') then
						(case log_activity_id when 1 then 'reading_' else 'writing_' end) ||
						(case unit_key when 'pages' then 'page' when 'sentences' then 'sentence' else 'character' end)
					else unit_key
				end || ':' || coalesce(language_code, '')
			)::uuid as stable_id
		from log_units
	)
	update log_units
	set id = stable.stable_id
	from stable
	where log_units.id = stable.id
		and log_units.id <> stable.stable_id`)
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
