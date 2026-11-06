package leaderboard

import (
	"context"
	"maps"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

var (
	versionsReader   = uuid.MustParse("aaaaaaaa-0000-4000-8000-000000000001")
	versionsOwner    = uuid.MustParse("aaaaaaaa-0000-4000-8000-000000000002")
	versionsContestC = uuid.MustParse("cccccccc-0000-4000-8000-000000000001")
	versionsContestD = uuid.MustParse("dddddddd-0000-4000-8000-000000000001")
	versionsLog1     = uuid.MustParse("11111111-0000-4000-8000-000000000001")
	versionsLog2     = uuid.MustParse("11111111-0000-4000-8000-000000000002")
	versionsLog3     = uuid.MustParse("11111111-0000-4000-8000-000000000003")
	versionsLog4     = uuid.MustParse("11111111-0000-4000-8000-000000000004")
	versionsLog5     = uuid.MustParse("11111111-0000-4000-8000-000000000005")
	versionsLog6     = uuid.MustParse("11111111-0000-4000-8000-000000000006")
)

const insertVersionsUserSQL = `insert into users (id, display_name) values ($1, 'Reader')`

const insertVersionsLogSQL = `insert into logs (id, user_id, language_code, log_activity_id, unit_key, amount, modifier,
	computed_score, eligible_official_leaderboard, created_at)
	values ($1, $2, 'jpn', 1, 'pages', 10, 1, 10, $3, $4)`

const insertVersionsContestLogSQL = `insert into contest_logs (contest_id, log_id, unit_key, amount, modifier,
	computed_score)
	values ($1, $2, 'pages', 10, 1, 10)`

func newVersionsDatabase(t *testing.T) *testpostgres.Database {
	t.Helper()

	db, err := testpostgres.New(t.Context())
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

func writeAs(t *testing.T, pool *pgxpool.Pool, key string, sql string, args ...any) {
	t.Helper()

	err := pgx.BeginFunc(t.Context(), pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(t.Context(), "select set_config('tadoku.tenant', $1, true)", key); err != nil {
			return err
		}
		_, err := tx.Exec(t.Context(), sql, args...)
		return err
	})
	if err != nil {
		t.Fatalf("write %q: %v", sql, err)
	}
}

func boardVersions(t *testing.T, db *testpostgres.Database, key string) map[string]int64 {
	t.Helper()

	rows, err := db.Pool.Query(t.Context(), `select board, version from leaderboard_versions where tenant = $1`, key)
	if err != nil {
		t.Fatal(err)
	}
	versions := map[string]int64{}
	var board string
	var version int64
	_, err = pgx.ForEachRow(rows, []any{&board, &version}, func() error {
		versions[board] = version
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return versions
}

func raisedBoards(before, after map[string]int64) map[string]bool {
	raised := map[string]bool{}
	for board, version := range after {
		if version > before[board] {
			raised[board] = true
		}
	}
	return raised
}

func seedVersionsBaseline(t *testing.T, db *testpostgres.Database, key string) {
	t.Helper()

	writeAs(t, db.AppPool, key, insertVersionsUserSQL, versionsReader)
	writeAs(t, db.AppPool, key, insertVersionsUserSQL, versionsOwner)
	for _, contest := range []uuid.UUID{versionsContestC, versionsContestD} {
		writeAs(t, db.AppPool, key, `insert into contests (id, owner_user_id, owner_user_display_name, private,
			contest_start, contest_end, registration_end, title, activity_type_id_allow_list, official)
			values ($1, $2, 'Owner', false, '2026-01-01', '2026-01-31', '2026-01-31', 'Round', array[1], false)`,
			contest,
			versionsOwner,
		)
	}
	writeAs(t, db.AppPool, key, `insert into contest_registrations (contest_id, user_id, language_codes)
		values ($1, $2, array['jpn'])`,
		versionsContestC,
		versionsReader,
	)

	for _, log := range []uuid.UUID{versionsLog1, versionsLog2, versionsLog3} {
		writeAs(t, db.AppPool, key, insertVersionsLogSQL, log, versionsReader, true, "2026-01-10")
		writeAs(t, db.AppPool, key, insertVersionsContestLogSQL, versionsContestC, log)
	}
	writeAs(t, db.AppPool, key, insertVersionsLogSQL, versionsLog4, versionsReader, false, "2026-01-10")
}

func TestLeaderboardVersionsRaiseExactlyTheBoardsAWriteChanges(t *testing.T) {
	t.Parallel()
	db := newVersionsDatabase(t)
	key := tenant.Production().String()
	seedVersionsBaseline(t, db, key)

	global := string(globalBoard)
	year2024, year2025, year2026 := string(yearlyBoard(2024)), string(yearlyBoard(2025)), string(yearlyBoard(2026))
	contestC, contestD := string(contestBoard(versionsContestC)), string(contestBoard(versionsContestD))
	tests := []struct {
		name string
		sql  string
		args []any
		want map[string]bool
	}{
		{
			name: "editing a description",
			sql:  `update logs set description = 'Edited' where id = $1`,
			args: []any{versionsLog1},
			want: map[string]bool{},
		},
		{
			name: "freezing a log",
			sql:  `update logs set frozen_at = now() where id = $1`,
			args: []any{versionsLog1},
			want: map[string]bool{},
		},
		{
			name: "changing the amount under a computed score",
			sql:  `update logs set amount = 20 where id = $1`,
			args: []any{versionsLog1},
			want: map[string]bool{},
		},
		{
			name: "changing a computed score",
			sql:  `update logs set computed_score = 12 where id = $1`,
			args: []any{versionsLog1},
			want: map[string]bool{global: true, year2026: true},
		},
		{
			name: "rescoring a contest log",
			sql:  `update contest_logs set computed_score = 12 where log_id = $1`,
			args: []any{versionsLog1},
			want: map[string]bool{contestC: true},
		},
		{
			name: "rescoring several logs in one statement",
			sql:  `update logs set computed_score = 15 where id = any($1)`,
			args: []any{[]uuid.UUID{versionsLog1, versionsLog2, versionsLog3}},
			want: map[string]bool{global: true, year2026: true},
		},
		{
			name: "creating an ineligible log",
			sql:  insertVersionsLogSQL,
			args: []any{versionsLog5, versionsReader, false, "2026-02-01"},
			want: map[string]bool{},
		},
		{
			name: "creating an eligible log",
			sql:  insertVersionsLogSQL,
			args: []any{versionsLog6, versionsReader, true, "2025-06-01"},
			want: map[string]bool{global: true, year2025: true},
		},
		{
			name: "moving a log to another year",
			sql:  `update logs set created_at = '2024-06-01' where id = $1`,
			args: []any{versionsLog6},
			want: map[string]bool{year2024: true, year2025: true},
		},
		{
			name: "making a log ineligible",
			sql:  `update logs set eligible_official_leaderboard = false where id = $1`,
			args: []any{versionsLog6},
			want: map[string]bool{global: true, year2024: true},
		},
		{
			name: "attaching a log to a contest",
			sql:  insertVersionsContestLogSQL,
			args: []any{versionsContestD, versionsLog4},
			want: map[string]bool{contestD: true},
		},
		{
			name: "moving a contest log to another contest",
			sql:  `update contest_logs set contest_id = $1 where log_id = $2`,
			args: []any{versionsContestC, versionsLog4},
			want: map[string]bool{contestC: true, contestD: true},
		},
		{
			name: "detaching a log from a contest",
			sql:  `delete from contest_logs where log_id = $1`,
			args: []any{versionsLog4},
			want: map[string]bool{contestC: true},
		},
		{
			name: "registering for a contest",
			sql:  `insert into contest_registrations (contest_id, user_id, language_codes) values ($1, $2, array['jpn'])`,
			args: []any{versionsContestD, versionsReader},
			want: map[string]bool{contestD: true},
		},
		{
			name: "changing registered languages",
			sql:  `update contest_registrations set language_codes = array['jpn', 'kor'] where contest_id = $1`,
			args: []any{versionsContestD},
			want: map[string]bool{},
		},
		{
			name: "leaving a contest",
			sql:  `update contest_registrations set deleted_at = now() where contest_id = $1`,
			args: []any{versionsContestD},
			want: map[string]bool{contestD: true},
		},
		{
			name: "purging a left registration",
			sql:  `delete from contest_registrations where contest_id = $1`,
			args: []any{versionsContestD},
			want: map[string]bool{},
		},
		{
			name: "retitling a contest",
			sql:  `update contests set title = 'Renamed' where id = $1`,
			args: []any{versionsContestC},
			want: map[string]bool{},
		},
		{
			name: "transferring a contest",
			sql:  `update contests set owner_user_id = $1 where id = $2`,
			args: []any{versionsReader, versionsContestD},
			want: map[string]bool{contestD: true},
		},
		{
			name: "deleting a contest",
			sql:  `update contests set deleted_at = now() where id = $1`,
			args: []any{versionsContestD},
			want: map[string]bool{contestD: true},
		},
		{
			name: "renaming a user",
			sql:  `update users set display_name = 'Renamed' where id = $1`,
			args: []any{versionsReader},
			want: map[string]bool{},
		},
		{
			name: "soft-deleting a log",
			sql:  `update logs set deleted_at = now() where id = $1`,
			args: []any{versionsLog2},
			want: map[string]bool{global: true, year2026: true, contestC: true},
		},
		{
			name: "hard-deleting a log",
			sql:  `delete from logs where id = $1`,
			args: []any{versionsLog3},
			want: map[string]bool{global: true, year2026: true, contestC: true},
		},
		{
			name: "deleting a user",
			sql:  `update users set deleted_at = now() where id = $1`,
			args: []any{versionsReader},
			want: map[string]bool{global: true, year2026: true},
		},
		{
			name: "purging a contest owner",
			sql:  `delete from users where id = $1`,
			args: []any{versionsOwner},
			want: map[string]bool{contestC: true},
		},
	}

	for _, test := range tests {
		before := boardVersions(t, db, key)

		writeAs(t, db.AppPool, key, test.sql, test.args...)

		if got := raisedBoards(before, boardVersions(t, db, key)); !maps.Equal(got, test.want) {
			t.Errorf("%s raised %v; want %v", test.name, got, test.want)
		}
	}
}

func TestLeaderboardVersionsStayInTheWritersTenant(t *testing.T) {
	t.Parallel()
	db := newVersionsDatabase(t)
	writer, bystander := "e2e/versions-0000000a", "e2e/versions-0000000b"
	_, err := db.Pool.Exec(
		t.Context(),
		`insert into tenants (key, kind) values ($1, 'test'), ($2, 'test')`,
		writer,
		bystander,
	)
	if err != nil {
		t.Fatal(err)
	}
	seedVersionsBaseline(t, db, bystander)
	before := boardVersions(t, db, bystander)

	writeAs(t, db.AppPool, writer, insertVersionsUserSQL, versionsReader)
	writeAs(t, db.AppPool, writer, insertVersionsLogSQL, versionsLog1, versionsReader, true, "2026-01-10")

	if got := boardVersions(t, db, writer); !maps.Equal(got, map[string]int64{"global": 1, "yearly:2026": 1}) {
		t.Errorf("writer versions = %v; want global and yearly:2026 at 1", got)
	}
	if got := boardVersions(t, db, bystander); !maps.Equal(got, before) {
		t.Errorf("bystander versions = %v; want unchanged %v", got, before)
	}
}

func TestLeaderboardVersionsFollowScenarioSeeds(t *testing.T) {
	t.Parallel()
	db := newVersionsDatabase(t)
	key, err := tenant.ParseTestTenant("e2e/versions-0000000c")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Pool.Exec(t.Context(), `insert into tenants (key, kind) values ($1, 'test')`, key.String())
	if err != nil {
		t.Fatal(err)
	}

	err = db.SeedTenant(t.Context(), key, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, insertVersionsUserSQL, versionsReader); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, insertVersionsLogSQL, versionsLog1, versionsReader, true, "2026-01-10")
		return err
	})
	if err != nil {
		t.Fatalf("seed as the scenario login: %v", err)
	}

	got := raisedBoards(map[string]int64{}, boardVersions(t, db, key.String()))
	if want := map[string]bool{"global": true, "yearly:2026": true}; !maps.Equal(got, want) {
		t.Errorf("seed raised %v; want %v", got, want)
	}
}

func TestLeaderboardVersionsAllowTestTenantDeletion(t *testing.T) {
	t.Parallel()
	db := newVersionsDatabase(t)
	key := "e2e/versions-0000000d"
	if _, err := db.Pool.Exec(t.Context(), `insert into tenants (key, kind) values ($1, 'test')`, key); err != nil {
		t.Fatal(err)
	}
	seedVersionsBaseline(t, db, key)

	_, err := db.Pool.Exec(t.Context(), `delete from tenants where key = $1`, key)

	if err != nil {
		t.Fatalf("delete test tenant: %v", err)
	}
	if got := boardVersions(t, db, key); len(got) != 0 {
		t.Errorf("versions after tenant deletion = %v; want none", got)
	}
}
