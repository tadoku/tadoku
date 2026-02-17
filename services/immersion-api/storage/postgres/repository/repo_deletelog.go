package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) DeleteLog(ctx context.Context, req *domain.LogDeleteRequest) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	isValid, err := qtx.CheckIfLogCanBeDeleted(ctx, postgres.CheckIfLogCanBeDeletedParams{
		Now:   req.Now,
		LogID: req.LogID,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not check if log can be deleted: %w", err)
	}

	if !isValid {
		_ = tx.Rollback()
		return domain.ErrForbidden
	}

	// Fetch outbox context before deleting
	logCtx, err := qtx.FetchLogOutboxContext(ctx, req.LogID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not fetch log context: %w", err)
	}

	contestIDs, err := qtx.FetchContestIDsForLog(ctx, req.LogID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not fetch contest IDs for log: %w", err)
	}

	if err := qtx.DeleteLog(ctx, req.LogID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not delete log: %w", err)
	}

	// Write outbox events for leaderboard sync
	if err = insertLeaderboardOutboxEvents(ctx, qtx, LeaderboardOutboxParams{
		UserID:          logCtx.UserID,
		ContestIDs:      contestIDs,
		OfficialContest: logCtx.EligibleOfficialLeaderboard,
		Year:            logCtx.Year,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not delete log: %w", err)
	}

	return nil
}
