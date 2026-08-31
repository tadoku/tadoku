package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/postgrestest"
)

func insertDeletionTestContest(t *testing.T, db *sql.DB, id, ownerID uuid.UUID, start, end time.Time, deletedAt *time.Time) {
	t.Helper()
	_, err := db.Exec(`
		insert into contests (
			id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
			registration_end, title, activity_type_id_allow_list, official, deleted_at
		) values ($1, $2, 'Identifying owner', false, $3, $4, $3, 'fixture', array[1], true, $5)
	`, id, ownerID, start, end, deletedAt)
	require.NoError(t, err)
}

func insertDeletionTestRegistration(t *testing.T, db *sql.DB, id, contestID, userID uuid.UUID) {
	t.Helper()
	_, err := db.Exec(`
		insert into contest_registrations (id, contest_id, user_id, language_codes)
		values ($1, $2, $3, array['eng'])
	`, id, contestID, userID)
	require.NoError(t, err)
}

func insertDeletionTestLog(t *testing.T, db *sql.DB, id, userID uuid.UUID, createdAt time.Time, score float64, description string, official bool) {
	t.Helper()
	_, err := db.Exec(`
		insert into logs (
			id, user_id, language_code, log_activity_id, duration_seconds, computed_score,
			eligible_official_leaderboard, description, created_at, updated_at
		) values ($1, $2, 'eng', 1, 60, $3, $4, $5, $6, $6)
	`, id, userID, score, official, description, createdAt)
	require.NoError(t, err)
	_, err = db.Exec("insert into log_tags (log_id, user_id, tag) values ($1, $2, 'private-tag')", id, userID)
	require.NoError(t, err)
}

func attachDeletionTestLog(t *testing.T, db *sql.DB, contestID, logID uuid.UUID, score float64) {
	t.Helper()
	_, err := db.Exec(`
		insert into contest_logs (contest_id, log_id, duration_seconds, computed_score)
		values ($1, $2, 60, $3)
	`, contestID, logID, score)
	require.NoError(t, err)
}

func scrubAccount(t *testing.T, repo *Repository, userID uuid.UUID, now time.Time) error {
	t.Helper()
	ctx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.ServiceIdentity{Name: "profile-api"})
	service := domain.NewAccountDeletionScrub(repo, commondomain.NewMockClock(now))
	return service.Execute(ctx, &domain.AccountDeletionScrubRequest{UserID: userID, RequestID: uuid.New()})
}

func TestAccountDeletionScrubPreservesHistoryAndRemovesNonHistoricalData(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	secondDeletedUserID := uuid.New()
	zeroScoreUserID := uuid.New()
	otherUserID := uuid.New()
	insertActiveUser(t, db, userID, "Identifying user")
	insertActiveUser(t, db, secondDeletedUserID, "Second identifying user")
	insertActiveUser(t, db, zeroScoreUserID, "Zero score user")
	insertActiveUser(t, db, otherUserID, "Other user")
	require.NoError(t, lockAccount(t, repo, userID, now.Add(-time.Minute)))
	require.NoError(t, lockAccount(t, repo, secondDeletedUserID, now.Add(-time.Minute)))

	completedContestID := uuid.New()
	openContestID := uuid.New()
	futureOwnedContestID := uuid.New()
	canceledContestID := uuid.New()
	canceledAt := now.Add(-24 * time.Hour)
	insertDeletionTestContest(t, db, completedContestID, otherUserID, now.AddDate(0, 0, -2), now.AddDate(0, 0, -1), nil)
	_, err := db.Exec("update contests set private = true where id = $1", completedContestID)
	require.NoError(t, err)
	insertDeletionTestContest(t, db, openContestID, otherUserID, now, now.AddDate(0, 0, 1), nil)
	insertDeletionTestContest(t, db, futureOwnedContestID, userID, now.AddDate(0, 0, 2), now.AddDate(0, 0, 3), nil)
	insertDeletionTestContest(t, db, canceledContestID, otherUserID, now.AddDate(0, 0, -4), now.AddDate(0, 0, -3), &canceledAt)

	completedRegistrationID := uuid.New()
	insertDeletionTestRegistration(t, db, completedRegistrationID, completedContestID, userID)
	insertDeletionTestRegistration(t, db, uuid.New(), completedContestID, secondDeletedUserID)
	insertDeletionTestRegistration(t, db, uuid.New(), completedContestID, zeroScoreUserID)
	insertDeletionTestRegistration(t, db, uuid.New(), openContestID, userID)
	insertDeletionTestRegistration(t, db, uuid.New(), futureOwnedContestID, userID)
	insertDeletionTestRegistration(t, db, uuid.New(), canceledContestID, userID)

	sharedLogID := uuid.New()
	completedLogID := uuid.New()
	openLogID := uuid.New()
	unlinkedLogID := uuid.New()
	secondDeletedLogID := uuid.New()
	insertDeletionTestLog(t, db, sharedLogID, userID, now.AddDate(0, 0, -2), 40, "shared private description", true)
	insertDeletionTestLog(t, db, completedLogID, userID, now.AddDate(-1, 0, 0), 60, "historical private description", true)
	insertDeletionTestLog(t, db, openLogID, userID, now, 20, "open private description", true)
	insertDeletionTestLog(t, db, unlinkedLogID, userID, now, 10, "unlinked private description", false)
	insertDeletionTestLog(t, db, secondDeletedLogID, secondDeletedUserID, now.AddDate(0, 0, -2), 100, "second private description", false)
	attachDeletionTestLog(t, db, completedContestID, sharedLogID, 40)
	attachDeletionTestLog(t, db, openContestID, sharedLogID, 41)
	attachDeletionTestLog(t, db, completedContestID, completedLogID, 60)
	attachDeletionTestLog(t, db, openContestID, openLogID, 20)
	attachDeletionTestLog(t, db, completedContestID, secondDeletedLogID, 100)
	var historicalLogUpdatedAtBefore time.Time
	var historicalContestUpdatedAtBefore time.Time
	var historicalScoreSnapshotBefore string
	require.NoError(t, db.QueryRow("select updated_at from logs where id = $1", sharedLogID).Scan(&historicalLogUpdatedAtBefore))
	require.NoError(t, db.QueryRow("select updated_at from contests where id = $1", completedContestID).Scan(&historicalContestUpdatedAtBefore))
	require.NoError(t, db.QueryRow(`
		select row_to_json(row(
			contest_id, log_id, unit_key, amount, modifier, duration_seconds, computed_score,
			score_rule_set_id, score_rule_ids, score_rates, score_source
		))::text
		from contest_logs where contest_id = $1 and log_id = $2
	`, completedContestID, sharedLogID).Scan(&historicalScoreSnapshotBefore))

	auditID := uuid.New()
	_, err = db.Exec(`
		insert into moderation_audit_log (id, user_id, action, metadata, description, created_at)
		values ($1, $2, 'reviewed', '{"private":"unchanged"}', 'immutable description', $3)
	`, auditID, userID, now.Add(-time.Hour))
	require.NoError(t, err)

	require.NoError(t, scrubAccount(t, repo, userID, now))
	require.NoError(t, scrubAccount(t, repo, userID, now.Add(time.Hour)))
	require.NoError(t, scrubAccount(t, repo, secondDeletedUserID, now))

	var displayName string
	var deletedAt time.Time
	require.NoError(t, db.QueryRow("select display_name, deleted_at from users where id = $1", userID).Scan(&displayName, &deletedAt))
	assert.Equal(t, "Deleted participant", displayName)
	assert.True(t, now.Equal(deletedAt))

	var retainedRegistrationUser uuid.UUID
	require.NoError(t, db.QueryRow("select user_id from contest_registrations where id = $1", completedRegistrationID).Scan(&retainedRegistrationUser))
	assert.Equal(t, userID, retainedRegistrationUser)
	var nonHistoricalRegistrations int
	require.NoError(t, db.QueryRow(`
		select count(*) from contest_registrations
		where user_id = $1 and contest_id <> $2
	`, userID, completedContestID).Scan(&nonHistoricalRegistrations))
	assert.Zero(t, nonHistoricalRegistrations)

	var futureDeletedAt time.Time
	var futureOwnerName string
	require.NoError(t, db.QueryRow("select deleted_at, owner_user_display_name from contests where id = $1", futureOwnedContestID).Scan(&futureDeletedAt, &futureOwnerName))
	assert.True(t, now.Equal(futureDeletedAt))
	assert.Equal(t, "Deleted organizer", futureOwnerName)

	for _, logID := range []uuid.UUID{sharedLogID, completedLogID} {
		var frozenAt time.Time
		var description sql.NullString
		require.NoError(t, db.QueryRow("select frozen_at, description from logs where id = $1", logID).Scan(&frozenAt, &description))
		assert.True(t, now.Equal(frozenAt))
		assert.False(t, description.Valid)
	}
	for _, logID := range []uuid.UUID{openLogID, unlinkedLogID} {
		var count int
		require.NoError(t, db.QueryRow("select count(*) from logs where id = $1", logID).Scan(&count))
		assert.Zero(t, count)
	}
	var tags int
	require.NoError(t, db.QueryRow("select count(*) from log_tags where user_id = $1", userID).Scan(&tags))
	assert.Zero(t, tags)

	var sharedLinks []uuid.UUID
	rows, err := db.Query("select contest_id from contest_logs where log_id = $1", sharedLogID)
	require.NoError(t, err)
	for rows.Next() {
		var id uuid.UUID
		require.NoError(t, rows.Scan(&id))
		sharedLinks = append(sharedLinks, id)
	}
	require.NoError(t, rows.Close())
	assert.Equal(t, []uuid.UUID{completedContestID}, sharedLinks)
	var retainedScore float64
	require.NoError(t, db.QueryRow("select computed_score from contest_logs where contest_id = $1 and log_id = $2", completedContestID, sharedLogID).Scan(&retainedScore))
	assert.Equal(t, 40.0, retainedScore)
	var historicalLogUpdatedAtAfter time.Time
	var historicalContestUpdatedAtAfter time.Time
	var historicalScoreSnapshotAfter string
	require.NoError(t, db.QueryRow("select updated_at from logs where id = $1", sharedLogID).Scan(&historicalLogUpdatedAtAfter))
	require.NoError(t, db.QueryRow("select updated_at from contests where id = $1", completedContestID).Scan(&historicalContestUpdatedAtAfter))
	require.NoError(t, db.QueryRow(`
		select row_to_json(row(
			contest_id, log_id, unit_key, amount, modifier, duration_seconds, computed_score,
			score_rule_set_id, score_rule_ids, score_rates, score_source
		))::text
		from contest_logs where contest_id = $1 and log_id = $2
	`, completedContestID, sharedLogID).Scan(&historicalScoreSnapshotAfter))
	assert.True(t, historicalLogUpdatedAtBefore.Equal(historicalLogUpdatedAtAfter))
	assert.True(t, historicalContestUpdatedAtBefore.Equal(historicalContestUpdatedAtAfter))
	assert.Equal(t, historicalScoreSnapshotBefore, historicalScoreSnapshotAfter)

	completedScores, err := repo.FetchAllContestLeaderboardScores(context.Background(), completedContestID)
	require.NoError(t, err)
	require.Len(t, completedScores, 3)
	completedScoresByUser := make(map[uuid.UUID]float64, len(completedScores))
	for _, entry := range completedScores {
		completedScoresByUser[entry.UserID] = entry.Score
	}
	assert.Equal(t, 100.0, completedScoresByUser[userID])
	assert.Equal(t, 100.0, completedScoresByUser[secondDeletedUserID])
	assert.Equal(t, 0.0, completedScoresByUser[zeroScoreUserID])
	openScores, err := repo.FetchAllContestLeaderboardScores(context.Background(), openContestID)
	require.NoError(t, err)
	assert.Empty(t, openScores)
	yearlyScores, err := repo.FetchAllYearlyLeaderboardScores(context.Background(), 2026)
	require.NoError(t, err)
	assert.Empty(t, yearlyScores)
	globalScores, err := repo.FetchAllGlobalLeaderboardScores(context.Background())
	require.NoError(t, err)
	assert.Empty(t, globalScores)
	leaderboard, err := repo.q.LeaderboardForContest(context.Background(), postgres.LeaderboardForContestParams{
		ContestID: completedContestID,
		PageSize:  10,
	})
	require.NoError(t, err)
	require.Len(t, leaderboard, 3)
	deletedEntries := 0
	zeroScoreEntries := 0
	for _, entry := range leaderboard {
		if entry.UserDisplayName == "Deleted participant" {
			deletedEntries++
			assert.Equal(t, int64(1), entry.Rank)
			assert.True(t, entry.IsTie)
			assert.Equal(t, float32(100), entry.Score)
		}
		if entry.UserID == zeroScoreUserID {
			zeroScoreEntries++
			assert.Equal(t, float32(0), entry.Score)
			assert.Equal(t, int64(3), entry.Rank)
		}
	}
	assert.Equal(t, 2, deletedEntries)
	assert.Equal(t, 1, zeroScoreEntries)

	var auditUserID uuid.UUID
	var auditMetadata string
	var auditDescription string
	require.NoError(t, db.QueryRow("select user_id, metadata::text, description from moderation_audit_log where id = $1", auditID).Scan(&auditUserID, &auditMetadata, &auditDescription))
	assert.Equal(t, userID, auditUserID)
	assert.JSONEq(t, `{"private":"unchanged"}`, auditMetadata)
	assert.Equal(t, "immutable description", auditDescription)

	var eventCount int
	require.NoError(t, db.QueryRow("select count(*) from leaderboard_outbox where user_id = $1", userID).Scan(&eventCount))
	assert.Equal(t, 5, eventCount)
	var completedRemovalEvents int
	require.NoError(t, db.QueryRow(`
		select count(*) from leaderboard_outbox
		where user_id = $1 and event_type = 'remove_contest_score' and contest_id = $2
	`, userID, completedContestID).Scan(&completedRemovalEvents))
	assert.Zero(t, completedRemovalEvents)
}

func TestAccountDeletionScrubRejectsRunningOwnedContestAndAcceptsExactEndBoundary(t *testing.T) {
	t.Run("running contest", func(t *testing.T) {
		db := postgrestest.OpenMigratedDatabase(t)
		repo := NewRepository(db)
		now := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		userID := uuid.New()
		insertActiveUser(t, db, userID, "owner")
		require.NoError(t, lockAccount(t, repo, userID, now.Add(-time.Minute)))
		insertDeletionTestContest(t, db, uuid.New(), userID, now, now, nil)

		assert.ErrorIs(t, scrubAccount(t, repo, userID, now), domain.ErrRunningContestOwned)
		var deletedAt sql.NullTime
		require.NoError(t, db.QueryRow("select deleted_at from users where id = $1", userID).Scan(&deletedAt))
		assert.False(t, deletedAt.Valid)
	})

	t.Run("midnight after contest end", func(t *testing.T) {
		db := postgrestest.OpenMigratedDatabase(t)
		repo := NewRepository(db)
		now := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		userID := uuid.New()
		insertActiveUser(t, db, userID, "owner")
		require.NoError(t, lockAccount(t, repo, userID, now.Add(-time.Minute)))
		contestID := uuid.New()
		insertDeletionTestContest(t, db, contestID, userID, now.AddDate(0, 0, -1), now.AddDate(0, 0, -1), nil)

		require.NoError(t, scrubAccount(t, repo, userID, now))
		var contestDeletedAt sql.NullTime
		require.NoError(t, db.QueryRow("select deleted_at from contests where id = $1", contestID).Scan(&contestDeletedAt))
		assert.False(t, contestDeletedAt.Valid)
		contest, err := repo.q.FindContestById(context.Background(), postgres.FindContestByIdParams{ID: contestID})
		require.NoError(t, err)
		assert.Equal(t, "Deleted organizer", contest.OwnerUserDisplayName)
	})
}

func TestFrozenLogRejectsMutationAndKeepsPrivateFieldsAndScore(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	logID := uuid.New()
	insertActiveUser(t, db, userID, "active user")
	insertDeletionTestLog(t, db, logID, userID, now.Add(-time.Hour), 25, "original description", false)
	_, err := db.Exec("update logs set frozen_at = $1 where id = $2", now, logID)
	require.NoError(t, err)

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	qtx := repo.q.WithTx(tx)
	require.NoError(t, lockUserForMutation(context.Background(), qtx, userID))
	assert.ErrorIs(t, lockLogForMutation(context.Background(), qtx, logID), domain.ErrLogFrozen)
	require.NoError(t, tx.Rollback())

	require.NoError(t, repo.q.UpdateLog(context.Background(), postgres.UpdateLogParams{
		DurationSeconds: sql.NullInt32{Int32: 120, Valid: true},
		ComputedScore:   sql.NullFloat64{Float64: 999, Valid: true},
		Description:     sql.NullString{String: "changed description", Valid: true},
		Now:             now.Add(time.Hour),
		LogID:           logID,
	}))

	var score float64
	var description string
	var tag string
	require.NoError(t, db.QueryRow("select computed_score, description from logs where id = $1", logID).Scan(&score, &description))
	require.NoError(t, db.QueryRow("select tag from log_tags where log_id = $1", logID).Scan(&tag))
	assert.Equal(t, 25.0, score)
	assert.Equal(t, "original description", description)
	assert.Equal(t, "private-tag", tag)
}
