package profile_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestProfileRepositoryListsEveryAcceptedDeletionStatus(t *testing.T) {
	t.Parallel()
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	acceptedAt := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	statuses := []string{
		"queued",
		"access_locked",
		"immersion_scrubbed",
		"caches_reconciled",
		"authorization_removed",
		"identity_deleted",
		"complete",
		"manual_attention",
	}
	want := make([]string, len(statuses))
	for i, status := range statuses {
		identityID := fmt.Sprintf("10000000-0000-4000-8000-%012d", i+1)
		want[i] = identityID
		if status == "manual_attention" {
			_, err = db.Pool.Exec(t.Context(), `
				insert into account_deletion_requests (
					identity_id, status, resume_status, accepted_at, discord_channel_id,
					discord_message_id, manual_attention_at, remediation_due_at
				) values (
					$1, $2, 'queued', $3::timestamp, 'channel', 'message',
					$3::timestamp, $3::timestamp + interval '7 days'
				)
			`, identityID, status, acceptedAt)
		} else {
			_, err = db.Pool.Exec(t.Context(), `
				insert into account_deletion_requests (
					identity_id, status, accepted_at, discord_channel_id, discord_message_id
				) values ($1, $2, $3, 'channel', 'message')
			`, identityID, status, acceptedAt)
		}
		if err != nil {
			t.Fatalf("insert %s: %v", status, err)
		}
	}
	if _, err := db.Pool.Exec(t.Context(), `
		insert into account_deletion_requests (identity_id, status, accepted_at)
		values ('20000000-0000-4000-8000-000000000000', 'receipt_pending', $1)
	`, acceptedAt); err != nil {
		t.Fatal(err)
	}

	got, err := profile.NewProfileRepository(db.Pool).ListAccountDeletionSuppressedIdentityIDs(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("suppressed identity IDs=%v, want %v", got, want)
	}
}
