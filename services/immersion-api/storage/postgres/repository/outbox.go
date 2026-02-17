package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

// LeaderboardOutboxParams describes which leaderboard outbox events to insert.
type LeaderboardOutboxParams struct {
	UserID          uuid.UUID
	ContestIDs      []uuid.UUID
	OfficialContest bool
	Year            int16
}

// insertLeaderboardOutboxEvents inserts outbox events for contest score refreshes
// and optionally official score refreshes within an existing transaction.
func insertLeaderboardOutboxEvents(ctx context.Context, qtx *postgres.Queries, p LeaderboardOutboxParams) error {
	for _, contestID := range p.ContestIDs {
		if err := qtx.InsertLeaderboardOutboxEvent(ctx, postgres.InsertLeaderboardOutboxEventParams{
			EventType: "refresh_contest_score",
			UserID:    p.UserID,
			ContestID: uuid.NullUUID{UUID: contestID, Valid: true},
		}); err != nil {
			return fmt.Errorf("could not insert contest score outbox event: %w", err)
		}
	}

	if p.OfficialContest {
		if err := qtx.InsertLeaderboardOutboxEvent(ctx, postgres.InsertLeaderboardOutboxEventParams{
			EventType: "refresh_official_scores",
			UserID:    p.UserID,
			Year:      sql.NullInt16{Int16: p.Year, Valid: true},
		}); err != nil {
			return fmt.Errorf("could not insert official scores outbox event: %w", err)
		}
	}

	return nil
}
