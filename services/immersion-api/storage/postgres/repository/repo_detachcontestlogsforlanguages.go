package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) DetachContestLogsForLanguages(ctx context.Context, req *domain.DetachContestLogsForLanguagesRequest) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	err = qtx.DetachContestLogsForLanguages(ctx, postgres.DetachContestLogsForLanguagesParams{
		ContestID:     req.ContestID,
		UserID:        req.UserID,
		LanguageCodes: req.LanguageCodes,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not detach contest logs for languages: %w", err)
	}

	// Write outbox events for leaderboard sync
	if err = qtx.InsertLeaderboardOutboxEvent(ctx, postgres.InsertLeaderboardOutboxEventParams{
		EventType: "refresh_contest_score",
		UserID:    req.UserID,
		ContestID: uuid.NullUUID{UUID: req.ContestID, Valid: true},
	}); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not insert outbox event: %w", err)
	}

	if req.OfficialContest {
		if err = qtx.InsertLeaderboardOutboxEvent(ctx, postgres.InsertLeaderboardOutboxEventParams{
			EventType: "refresh_official_scores",
			UserID:    req.UserID,
			Year:      sql.NullInt16{Int16: req.Year, Valid: true},
		}); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not insert outbox event: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not detach contest logs for languages: %w", err)
	}

	return nil
}
