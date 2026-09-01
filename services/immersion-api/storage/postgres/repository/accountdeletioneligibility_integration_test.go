package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/postgrestest"
)

func TestAccountDeletionEligibilityUsesLatestRunningContestCompletion(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	checkedAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	insertActiveUser(t, db, userID, "contest owner")

	insertDeletionTestContest(t, db, uuid.New(), userID, checkedAt.AddDate(0, 0, -1), checkedAt.AddDate(0, 0, 1), nil)
	insertDeletionTestContest(t, db, uuid.New(), userID, checkedAt, checkedAt.AddDate(0, 0, 3), nil)
	insertDeletionTestContest(t, db, uuid.New(), userID, checkedAt.AddDate(0, 0, 1), checkedAt.AddDate(0, 0, 5), nil)
	canceledAt := checkedAt.Add(-time.Hour)
	insertDeletionTestContest(t, db, uuid.New(), userID, checkedAt.AddDate(0, 0, -1), checkedAt.AddDate(0, 0, 10), &canceledAt)

	availableAfter, err := repo.FindRunningOwnedContestAvailableAfter(context.Background(), userID, checkedAt)
	require.NoError(t, err)
	require.NotNil(t, availableAfter)
	assert.Equal(t, time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC), *availableAfter)
}

func TestAccountDeletionEligibilityAcceptsExactCompletionBoundary(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	checkedAt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	userID := uuid.New()
	insertActiveUser(t, db, userID, "contest owner")
	insertDeletionTestContest(t, db, uuid.New(), userID, checkedAt.AddDate(0, 0, -2), checkedAt.AddDate(0, 0, -1), nil)

	availableAfter, err := repo.FindRunningOwnedContestAvailableAfter(context.Background(), userID, checkedAt)
	require.NoError(t, err)
	assert.Nil(t, availableAfter)
}

func TestAccountDeletionLockRejectsRunningOwnerWithoutPersistingLock(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	checkedAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	insertActiveUser(t, db, userID, "contest owner")
	insertDeletionTestContest(t, db, uuid.New(), userID, checkedAt, checkedAt.AddDate(0, 0, 1), nil)

	assert.ErrorIs(t, lockAccount(t, repo, userID, checkedAt), domain.ErrRunningContestOwned)
	var lockedAt sql.NullTime
	require.NoError(t, db.QueryRow("select deletion_locked_at from users where id = $1", userID).Scan(&lockedAt))
	assert.False(t, lockedAt.Valid)
}

func TestAccountDeletionLockWaitsForContestCreationAndRejects(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	checkedAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	insertActiveUser(t, db, userID, "contest owner")

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	require.NoError(t, lockUserForMutation(context.Background(), repo.q.WithTx(tx), userID))
	_, err = tx.Exec(`
		insert into contests (
			id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
			registration_end, title, activity_type_id_allow_list, official
		) values ($1, $2, 'contest owner', false, $3, $4, $3, 'racing contest', array[1], false)
	`, uuid.New(), userID, checkedAt, checkedAt.AddDate(0, 0, 1))
	require.NoError(t, err)

	lockResult := make(chan error, 1)
	go func() { lockResult <- lockAccount(t, repo, userID, checkedAt) }()
	select {
	case err := <-lockResult:
		require.Failf(t, "deletion lock did not wait", "returned early with %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	require.NoError(t, tx.Commit())
	assert.ErrorIs(t, <-lockResult, domain.ErrRunningContestOwned)
}
