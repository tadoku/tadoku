package leaderboard

import (
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestCleanupOutboxRetainsPendingAndRecentEvents(t *testing.T) {
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	_, err = db.Pool.Exec(t.Context(), `
		insert into leaderboard_outbox (event_type, user_id, processed_at) values
			('old_processed', '11111111-1111-4111-8111-111111111111', '2026-09-20'),
			('recent_processed', '11111111-1111-4111-8111-111111111111', '2026-09-22'),
			('old_pending', '11111111-1111-4111-8111-111111111111', null)
	`)
	if err != nil {
		t.Fatal(err)
	}
	cutoff := time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC)
	if err := NewRepository(db.Pool).cleanupOutbox(t.Context(), cutoff); err != nil {
		t.Fatal(err)
	}

	rows, err := db.Pool.Query(t.Context(), `select event_type from leaderboard_outbox order by event_type`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var eventType string
		if err := rows.Scan(&eventType); err != nil {
			t.Fatal(err)
		}
		got = append(got, eventType)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "old_pending" || got[1] != "recent_processed" {
		t.Errorf("remaining events = %v, want old_pending and recent_processed", got)
	}
}
