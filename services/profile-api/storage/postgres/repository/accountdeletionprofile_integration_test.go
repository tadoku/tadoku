package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/profile-api/storage/postgres/postgrestest"
)

func TestDeleteProfileIsIdempotentAgainstMigratedSchema(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	identityID := uuid.New()
	_, err := db.Exec("insert into profiles (user_id) values ($1)", identityID)
	require.NoError(t, err)

	require.NoError(t, repo.DeleteProfile(context.Background(), identityID))
	require.NoError(t, repo.DeleteProfile(context.Background(), identityID))

	var count int
	require.NoError(t, db.QueryRow("select count(*) from profiles where user_id = $1", identityID).Scan(&count))
	assert.Zero(t, count)
}

func TestListAccountDeletionSuppressedIdentityIDsUsesAcceptedStates(t *testing.T) {
	db := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(db)
	acceptedAt := time.Date(2026, 9, 1, 15, 0, 0, 0, time.UTC)
	acceptedStatuses := []string{
		"queued",
		"access_locked",
		"immersion_scrubbed",
		"caches_reconciled",
		"authorization_removed",
		"identity_deleted",
		"complete",
	}
	want := make([]uuid.UUID, 0, len(acceptedStatuses)+1)
	for _, status := range acceptedStatuses {
		identityID := uuid.New()
		want = append(want, identityID)
		insertDeletionRequest(t, db, identityID, status, acceptedAt)
	}
	manualAttentionID := uuid.New()
	want = append(want, manualAttentionID)
	_, err := db.Exec(`
		insert into account_deletion_requests (
			identity_id, status, resume_status, accepted_at, discord_channel_id,
			discord_message_id, manual_attention_at, remediation_due_at
		) values (
			$1, 'manual_attention', 'queued', $2::timestamp, 'channel', 'message',
			$2::timestamp, $2::timestamp + interval '7 days'
		)
	`, manualAttentionID, acceptedAt)
	require.NoError(t, err)
	insertDeletionRequest(t, db, uuid.New(), "receipt_pending", acceptedAt)

	got, err := repo.ListAccountDeletionSuppressedIdentityIDs(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, want, got)
}

func insertDeletionRequest(t *testing.T, db *sql.DB, identityID uuid.UUID, status string, acceptedAt time.Time) {
	t.Helper()
	if status == "receipt_pending" {
		_, err := db.Exec(`
			insert into account_deletion_requests (identity_id, status, accepted_at)
			values ($1, $2, $3)
		`, identityID, status, acceptedAt)
		require.NoError(t, err)
		return
	}
	_, err := db.Exec(`
		insert into account_deletion_requests (
			identity_id, status, accepted_at, discord_channel_id, discord_message_id
		) values ($1, $2, $3, 'channel', 'message')
	`, identityID, status, acceptedAt)
	require.NoError(t, err)
}
