package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) UpdateLog(ctx context.Context, req *domain.LogUpdateRequest) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	// Fetch outbox context before changes
	logCtx, err := qtx.FetchLogOutboxContext(ctx, req.LogID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not fetch log context: %w", err)
	}

	tracking := req.Tracking()

	// Update the log itself
	if err := qtx.UpdateLog(ctx, postgres.UpdateLogParams{
		LogID:           req.LogID,
		Amount:          trackingAmount(tracking),
		Modifier:        trackingModifier(tracking),
		UnitID:          trackingUnitID(tracking),
		DurationSeconds: trackingDurationSeconds(tracking),
		ComputedScore:   postgres.NewNullFloat64FromFloat32(tracking.ComputedScore),
		Description:     postgres.NewNullString(req.Description),
		Now:             req.Now(),
	}); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not update log: %w", err)
	}

	// Update contest_logs for ongoing contests only
	if err := qtx.UpdateOngoingContestLogs(ctx, postgres.UpdateOngoingContestLogsParams{
		LogID:           req.LogID,
		Amount:          trackingAmount(tracking),
		Modifier:        trackingModifier(tracking),
		DurationSeconds: trackingDurationSeconds(tracking),
		ComputedScore:   postgres.NewNullFloat64FromFloat32(tracking.ComputedScore),
		Now:             req.Now(),
	}); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not update contest logs: %w", err)
	}

	// Sync tags: delete all, re-insert new ones
	if err := qtx.DeleteLogTagsForLog(ctx, req.LogID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not delete old tags: %w", err)
	}
	for _, tag := range req.Tags {
		if err := qtx.InsertLogTag(ctx, postgres.InsertLogTagParams{
			LogID:  req.LogID,
			UserID: req.UserID(),
			Tag:    tag,
		}); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not insert log tag: %w", err)
		}
	}

	// Emit outbox events for ongoing contests
	ongoingContestIDs, err := qtx.FetchOngoingContestIDsForLog(ctx, postgres.FetchOngoingContestIDsForLogParams{
		LogID: req.LogID,
		Now:   req.Now(),
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not fetch ongoing contest IDs: %w", err)
	}

	if err = insertLeaderboardOutboxEvents(ctx, qtx, LeaderboardOutboxParams{
		UserID:          logCtx.UserID,
		ContestIDs:      ongoingContestIDs,
		OfficialContest: logCtx.EligibleOfficialLeaderboard,
		Year:            logCtx.Year,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}
