package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) UpdateLog(ctx context.Context, req *domain.LogUpdateRequest) error {
	activity := req.Activity()

	// Look up the existing log to get activity_id + language_code for unit validation
	existingLog, err := r.q.FindLogByID(ctx, postgres.FindLogByIDParams{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("could not fetch log: %w", err)
	}

	// Resolve unit -> modifier (only when unit is provided)
	var modifier *float32
	if req.UnitID != nil {
		unit, err := r.q.FindUnitForTracking(ctx, postgres.FindUnitForTrackingParams{
			ID:            *req.UnitID,
			LogActivityID: existingLog.ActivityID,
			LanguageCode:  postgres.NewNullString(&existingLog.LanguageCode),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("invalid unit supplied: %w", domain.ErrInvalidLog)
			}
			return fmt.Errorf("could not fetch unit for tracking: %w", err)
		}
		modifier = &unit.Modifier
	}

	// Compute score using resolved modifier
	computedScore := domain.ComputeScore(activity, req.Amount, modifier, req.DurationSeconds)

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

	// Update the log itself
	if err := qtx.UpdateLog(ctx, postgres.UpdateLogParams{
		LogID:           req.LogID,
		Amount:          postgres.NewNullFloat64(req.Amount),
		Modifier:        postgres.NewNullFloat64(modifier),
		UnitID:          postgres.NewNullUUID(req.UnitID),
		DurationSeconds: postgres.NewNullInt32(req.DurationSeconds),
		ComputedScore:   sql.NullFloat64{Valid: true, Float64: float64(computedScore)},
		Description:     postgres.NewNullString(req.Description),
		Now:             req.Now(),
	}); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not update log: %w", err)
	}

	// Update contest_logs for ongoing contests only
	if err := qtx.UpdateOngoingContestLogs(ctx, postgres.UpdateOngoingContestLogsParams{
		LogID:           req.LogID,
		Amount:          postgres.NewNullFloat64(req.Amount),
		Modifier:        postgres.NewNullFloat64(modifier),
		DurationSeconds: postgres.NewNullInt32(req.DurationSeconds),
		ComputedScore:   sql.NullFloat64{Valid: true, Float64: float64(computedScore)},
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
