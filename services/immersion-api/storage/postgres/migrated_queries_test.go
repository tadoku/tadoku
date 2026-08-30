package postgres_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/postgrestest"
)

func TestJoinedQueriesAgainstMigratedSchema(t *testing.T) {
	testDB := postgrestest.OpenMigratedDatabase(t)
	ctx := context.Background()
	userID := uuid.New()
	contestID := uuid.New()
	logID := uuid.New()

	seedMigratedQueryFixture(t, testDB, userID, contestID, logID)
	queries := postgres.NewQueries(testDB)

	t.Run("FindLogByID", func(t *testing.T) {
		log, err := queries.FindLogByID(ctx, postgres.FindLogByIDParams{
			ID:             logID,
			IncludeDeleted: false,
		})
		require.NoError(t, err)
		assert.Equal(t, userID, log.UserID)
		assert.Equal(t, "Regression User", log.UserDisplayName)
	})

	t.Run("ListContests", func(t *testing.T) {
		contests, err := queries.ListContests(ctx, postgres.ListContestsParams{
			IncludeDeleted: false,
			Official:       false,
			IncludePrivate: true,
			PageSize:       10,
		})
		require.NoError(t, err)
		require.Len(t, contests, 1)
		assert.Equal(t, contestID, contests[0].ID)
		assert.Equal(t, "Regression User", contests[0].OwnerUserDisplayName)
	})

	t.Run("FindContestByID", func(t *testing.T) {
		contest, err := queries.FindContestById(ctx, postgres.FindContestByIdParams{
			ID:             contestID,
			IncludeDeleted: false,
		})
		require.NoError(t, err)
		assert.Equal(t, contestID, contest.ID)
		assert.Equal(t, "Regression User", contest.OwnerUserDisplayName)
	})

	t.Run("LeaderboardForContest", func(t *testing.T) {
		leaderboard, err := queries.LeaderboardForContest(ctx, postgres.LeaderboardForContestParams{
			ContestID: contestID,
			PageSize:  10,
		})
		require.NoError(t, err)
		require.Len(t, leaderboard, 1)
		assert.Equal(t, userID, leaderboard[0].UserID)
		assert.Equal(t, float32(12), leaderboard[0].Score)
	})
}

func seedMigratedQueryFixture(t testing.TB, testDB *sql.DB, userID, contestID, logID uuid.UUID) {
	t.Helper()

	tx, err := testDB.Begin()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tx.Rollback()
	})

	_, err = tx.Exec(`insert into users (id, display_name)
values ($1, 'Regression User')`, userID)
	require.NoError(t, err)

	_, err = tx.Exec(`
insert into contests (
  id,
  owner_user_id,
  owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  language_code_allow_list,
  activity_type_id_allow_list,
  official
) values (
  $1,
  $2,
  'Regression User',
  false,
  date '2026-08-01',
  date '2026-08-31',
  date '2026-08-31',
  'Migration regression contest',
  array['eng']::varchar(10)[],
  array[1]::integer[],
  false
)`, contestID, userID)
	require.NoError(t, err)

	_, err = tx.Exec(`insert into contest_registrations (id, contest_id, user_id, language_codes)
values ($1, $2, $3, array['eng']::varchar(10)[])`, uuid.New(), contestID, userID)
	require.NoError(t, err)

	_, err = tx.Exec(`
insert into logs (
  id,
  user_id,
  language_code,
  log_activity_id,
  duration_seconds,
  computed_score,
  eligible_official_leaderboard,
  description
) values (
  $1,
  $2,
  'eng',
  1,
  3600,
  12,
  true,
  'Migration regression log'
)`, logID, userID)
	require.NoError(t, err)

	_, err = tx.Exec(`
insert into contest_logs (
  contest_id,
  log_id,
  duration_seconds,
  computed_score
) values (
  $1,
  $2,
  3600,
  12
);`, contestID, logID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
}
