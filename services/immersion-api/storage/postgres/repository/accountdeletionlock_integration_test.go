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

func lockAccount(t *testing.T, repo *Repository, userID uuid.UUID, lockedAt time.Time) error {
	t.Helper()
	ctx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.ServiceIdentity{
		Subject:   "system:serviceaccount:tadoku:profile-api",
		Name:      "profile-api",
		Namespace: "tadoku",
		Audience:  []string{"immersion-api"},
	})
	service := domain.NewAccountDeletionLock(repo, commondomain.NewMockClock(lockedAt))
	return service.Execute(ctx, &domain.AccountDeletionLockRequest{
		UserID:    userID,
		RequestID: uuid.New(),
	})
}

func insertActiveUser(t *testing.T, db *sql.DB, userID uuid.UUID, displayName string) {
	t.Helper()
	_, err := db.Exec("insert into users (id, display_name) values ($1, $2)", userID, displayName)
	require.NoError(t, err)
}

func TestAccountDeletionLockWaitsForEarlierMutation(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	userID := uuid.New()
	insertActiveUser(t, db, userID, "before")

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	user, err := repo.q.WithTx(tx).LockUserForMutation(context.Background(), userID)
	require.NoError(t, err)
	require.False(t, user.DeletionLockedAt.Valid)
	require.False(t, user.DeletedAt.Valid)

	lockedAt := time.Date(2026, 8, 29, 17, 0, 0, 0, time.UTC)
	lockResult := make(chan error, 1)
	go func() {
		lockResult <- lockAccount(t, repo, userID, lockedAt)
	}()

	select {
	case err := <-lockResult:
		require.Failf(t, "deletion lock did not wait", "returned early with %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	_, err = tx.Exec("update users set display_name = 'mutation-won' where id = $1", userID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, <-lockResult)

	var displayName string
	var actualLockedAt time.Time
	require.NoError(t, db.QueryRow(
		"select display_name, deletion_locked_at from users where id = $1",
		userID,
	).Scan(&displayName, &actualLockedAt))
	assert.Equal(t, "mutation-won", displayName)
	assert.True(t, lockedAt.Equal(actualLockedAt))
}

func TestMutationRejectsAfterDeletionLock(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	userID := uuid.New()
	insertActiveUser(t, db, userID, "unchanged")
	require.NoError(t, lockAccount(t, repo, userID, time.Now().UTC()))

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	err = lockUserForMutation(context.Background(), repo.q.WithTx(tx), userID)
	assert.ErrorIs(t, err, domain.ErrAccountDeletionInProgress)
	require.NoError(t, tx.Rollback())

	var displayName string
	require.NoError(t, db.QueryRow("select display_name from users where id = $1", userID).Scan(&displayName))
	assert.Equal(t, "unchanged", displayName)
}

func TestDeletionLockCreatesTombstoneBeforeFirstImmersionWrite(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	userID := uuid.New()
	lockedAt := time.Date(2026, 8, 29, 18, 0, 0, 0, time.UTC)

	require.NoError(t, lockAccount(t, repo, userID, lockedAt))
	require.NoError(t, lockAccount(t, repo, userID, lockedAt.Add(time.Hour)))

	_, err := repo.q.UpsertUser(context.Background(), postgres.UpsertUserParams{
		ID:               userID,
		DisplayName:      "must-not-return",
		SessionCreatedAt: lockedAt.Add(time.Hour),
	})
	assert.ErrorIs(t, err, sql.ErrNoRows)

	var displayName string
	var actualLockedAt time.Time
	require.NoError(t, db.QueryRow(
		"select display_name, deletion_locked_at from users where id = $1",
		userID,
	).Scan(&displayName, &actualLockedAt))
	assert.Empty(t, displayName)
	assert.True(t, lockedAt.Equal(actualLockedAt))
}
