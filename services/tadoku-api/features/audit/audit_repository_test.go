package audit

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestRepositoryPersistsAudit(t *testing.T) {
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

	event := Event{
		ActorID: uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		Action:  "ban_user",
		Metadata: map[string]any{
			"target_user_id": "11111111-1111-4111-8111-111111111111",
			"new_role":       "banned",
		},
		Description: "repeated abuse",
		recordedAt:  time.Date(2026, 9, 19, 12, 34, 56, 0, time.UTC),
	}
	repository := NewRepository(db.Pool)
	if err := repository.Create(t.Context(), event); err != nil {
		t.Fatal(err)
	}

	var (
		actorID     uuid.UUID
		action      string
		metadataRaw []byte
		description string
		recordedAt  time.Time
	)
	err = db.Pool.QueryRow(t.Context(), `
		select user_id, action, metadata, description, created_at
		from moderation_audit_log
	`).Scan(&actorID, &action, &metadataRaw, &description, &recordedAt)
	if err != nil {
		t.Fatal(err)
	}

	var metadata map[string]any
	if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
		t.Fatal(err)
	}
	if actorID != event.ActorID {
		t.Errorf("actor ID = %s, want %s", actorID, event.ActorID)
	}
	if action != event.Action {
		t.Errorf("action = %q, want %q", action, event.Action)
	}
	if !reflect.DeepEqual(metadata, event.Metadata) {
		t.Errorf("metadata = %v, want %v", metadata, event.Metadata)
	}
	if description != event.Description {
		t.Errorf("description = %q, want %q", description, event.Description)
	}
	if !recordedAt.Equal(event.recordedAt) {
		t.Errorf("recorded at = %s, want %s", recordedAt, event.recordedAt)
	}
}
