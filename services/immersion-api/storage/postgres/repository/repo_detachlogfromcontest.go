package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) DetachLogFromContest(ctx context.Context, req *domain.ContestModerationDetachLogRequest, moderatorUserID uuid.UUID) error {
	// Start transaction
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	// Create audit log entry
	metadata := map[string]interface{}{
		"contest_id": req.ContestID.String(),
		"log_id":     req.LogID.String(),
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not marshal metadata: %w", err)
	}

	err = qtx.CreateModerationAuditLog(ctx, postgres.CreateModerationAuditLogParams{
		UserID:      moderatorUserID,
		Action:      "detach_log",
		Metadata:    metadataJSON,
		Description: postgres.NewNullString(&req.Reason),
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not create audit log: %w", err)
	}

	// Look up the log owner for the outbox event
	logCtx, err := qtx.FetchLogOutboxContext(ctx, req.LogID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not fetch log context: %w", err)
	}

	// Detach log from contest
	err = qtx.DetachLogFromContest(ctx, postgres.DetachLogFromContestParams{
		ContestID: req.ContestID,
		LogID:     req.LogID,
	})
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("could not detach log from contest: %w", err)
	}

	// Write outbox events for leaderboard sync
	if err = qtx.InsertLeaderboardOutboxEvent(ctx, postgres.InsertLeaderboardOutboxEventParams{
		EventType: "refresh_contest_score",
		UserID:    logCtx.UserID,
		ContestID: uuid.NullUUID{UUID: req.ContestID, Valid: true},
	}); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not insert outbox event: %w", err)
	}

	if logCtx.EligibleOfficialLeaderboard {
		if err = qtx.InsertLeaderboardOutboxEvent(ctx, postgres.InsertLeaderboardOutboxEventParams{
			EventType: "refresh_official_scores",
			UserID:    logCtx.UserID,
			Year:      sql.NullInt16{Int16: logCtx.Year, Valid: true},
		}); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not insert outbox event: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}
