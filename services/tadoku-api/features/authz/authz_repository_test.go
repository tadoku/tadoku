package authz_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestAuthzRepositoryPersistsModerationAudit(t *testing.T) {
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

	audit := authz.ModerationAudit{
		ModeratorUserID: uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		Action:          authz.ModerationActionBanUser,
		TargetUserID:    uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		NewRole:         authz.RoleBanned,
		Description:     "repeated abuse",
		CreatedAt:       time.Date(2026, 9, 19, 12, 34, 56, 0, time.UTC),
	}
	repository := authz.NewAuthzRepository(db.Pool)
	if err := repository.CreateModerationAudit(t.Context(), audit); err != nil {
		t.Fatal(err)
	}

	var (
		moderatorID uuid.UUID
		action      string
		metadataRaw []byte
		description string
		createdAt   time.Time
	)
	err = db.Pool.QueryRow(t.Context(), `
		select user_id, action, metadata, description, created_at
		from moderation_audit_log
	`).Scan(&moderatorID, &action, &metadataRaw, &description, &createdAt)
	if err != nil {
		t.Fatal(err)
	}

	var metadata map[string]string
	if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
		t.Fatal(err)
	}
	wantMetadata := map[string]string{
		"target_user_id": audit.TargetUserID.String(),
		"new_role":       string(audit.NewRole),
	}
	if moderatorID != audit.ModeratorUserID {
		t.Errorf("moderator ID = %s, want %s", moderatorID, audit.ModeratorUserID)
	}
	if action != string(audit.Action) {
		t.Errorf("action = %q, want %q", action, audit.Action)
	}
	if !reflect.DeepEqual(metadata, wantMetadata) {
		t.Errorf("metadata = %v, want %v", metadata, wantMetadata)
	}
	if description != audit.Description {
		t.Errorf("description = %q, want %q", description, audit.Description)
	}
	if !createdAt.Equal(audit.CreatedAt) {
		t.Errorf("created at = %s, want %s", createdAt, audit.CreatedAt)
	}
}
