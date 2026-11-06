package leaderboard

import (
	"context"
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

const missingVersionTriggersSQL = `select tables.name || ' ' || events.name
	from unnest($1::text[]) tables (name)
	cross join (values ('insert', 4), ('update', 16), ('delete', 8)) events (name, bit)
	where not exists (
		select 1
		from pg_trigger
		inner join pg_proc on pg_proc.oid = pg_trigger.tgfoid
		where pg_trigger.tgrelid = to_regclass(tables.name)
			and not pg_trigger.tgisinternal
			and pg_trigger.tgenabled <> 'D'
			and pg_trigger.tgtype & (1 | 2 | 64) = 1
			and pg_trigger.tgtype & events.bit <> 0
			and pg_trigger.tgdeferrable
			and pg_trigger.tginitdeferred
			and pg_proc.proname = 'leaderboard_versions_from_' || tables.name
	)
	order by 1`

func leaderboardInputs(t *testing.T) []string {
	t.Helper()

	path, err := bazel.Runfile("services/tadoku-api/sql/leaderboard/leaderboard.sql")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := pg_query.ParseToJSON(string(source))
	if err != nil {
		t.Fatal(err)
	}
	var tree any
	if err := json.Unmarshal([]byte(parsed), &tree); err != nil {
		t.Fatal(err)
	}

	var relations, commonTables []string
	var walk func(any)
	walk = func(node any) {
		switch node := node.(type) {
		case map[string]any:
			if rangeVar, ok := node["RangeVar"].(map[string]any); ok {
				relations = append(relations, rangeVar["relname"].(string))
			}
			if commonTable, ok := node["CommonTableExpr"].(map[string]any); ok {
				commonTables = append(commonTables, commonTable["ctename"].(string))
			}
			for _, child := range node {
				walk(child)
			}
		case []any:
			for _, child := range node {
				walk(child)
			}
		}
	}
	walk(tree)

	relations = slices.DeleteFunc(relations, func(name string) bool {
		return name == "leaderboard_versions" || slices.Contains(commonTables, name)
	})
	slices.Sort(relations)
	return slices.Compact(relations)
}

func missingVersionTriggers(ctx context.Context, pool *pgxpool.Pool, tables []string) ([]string, error) {
	rows, err := pool.Query(ctx, missingVersionTriggersSQL, tables)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

func TestLeaderboardVersionsCoverEveryLeaderboardInput(t *testing.T) {
	t.Parallel()
	db := newVersionsDatabase(t)
	inputs := leaderboardInputs(t)

	missing, err := missingVersionTriggers(t.Context(), db.Pool, inputs)

	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 || len(missing) != 0 {
		t.Errorf("leaderboard inputs %v lack version triggers for %v", inputs, missing)
	}
}

func TestLeaderboardVersionCoverageReportsADroppedTrigger(t *testing.T) {
	t.Parallel()
	db := newVersionsDatabase(t)
	if _, err := db.Pool.Exec(t.Context(), `drop trigger leaderboard_versions_logs on logs`); err != nil {
		t.Fatal(err)
	}

	missing, err := missingVersionTriggers(t.Context(), db.Pool, leaderboardInputs(t))

	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"logs delete", "logs insert", "logs update"}; !slices.Equal(missing, want) {
		t.Errorf("missing triggers = %v; want %v", missing, want)
	}
}
