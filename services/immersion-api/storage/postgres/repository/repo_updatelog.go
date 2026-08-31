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
	if err = lockUserForMutation(ctx, qtx, req.UserID()); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = lockLogForMutation(ctx, qtx, req.LogID); err != nil {
		_ = tx.Rollback()
		return err
	}

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
		UnitKey:         trackingUnitKey(tracking),
		DurationSeconds: trackingDurationSeconds(tracking),
		ComputedScore:   postgres.NewNullFloat64FromFloat32(tracking.ComputedScore),
		ScoreRuleSetID:  scoreRuleSetID(tracking.ScoreProvenance),
		ScoreRuleIds:    scoreRuleIDs(tracking.ScoreProvenance),
		ScoreRates:      scoreRates(tracking.ScoreProvenance),
		ScoreSource:     scoreSource(tracking.ScoreProvenance),
		Description:     postgres.NewNullString(req.Description),
		Now:             req.Now(),
	}); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not update log: %w", err)
	}

	contestTrackings := req.ContestTrackings()
	if len(contestTrackings) == 0 {
		if err := qtx.UpdateOngoingContestLogs(ctx, postgres.UpdateOngoingContestLogsParams{
			LogID:           req.LogID,
			UnitKey:         trackingUnitKey(tracking),
			Amount:          trackingAmount(tracking),
			Modifier:        trackingModifier(tracking),
			DurationSeconds: trackingDurationSeconds(tracking),
			ComputedScore:   postgres.NewNullFloat64FromFloat32(tracking.ComputedScore),
			ScoreRuleSetID:  scoreRuleSetID(nil),
			ScoreRuleIds:    scoreRuleIDs(nil),
			ScoreRates:      scoreRates(nil),
			ScoreSource:     scoreSource(nil),
			Now:             req.Now(),
		}); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not update contest logs: %w", err)
		}
	} else {
		for _, contestTracking := range contestTrackings {
			contestTracking := contestTracking
			if err := qtx.UpdateOngoingContestLog(ctx, postgres.UpdateOngoingContestLogParams{
				LogID:           req.LogID,
				ContestID:       contestTracking.ContestID,
				UnitKey:         trackingUnitKey(contestTracking.Tracking),
				Amount:          trackingAmount(contestTracking.Tracking),
				Modifier:        trackingModifier(contestTracking.Tracking),
				DurationSeconds: trackingDurationSeconds(contestTracking.Tracking),
				ComputedScore:   postgres.NewNullFloat64FromFloat32(contestTracking.Tracking.ComputedScore),
				ScoreRuleSetID:  scoreRuleSetID(contestTracking.Tracking.ScoreProvenance),
				ScoreRuleIds:    scoreRuleIDs(contestTracking.Tracking.ScoreProvenance),
				ScoreRates:      scoreRates(contestTracking.Tracking.ScoreProvenance),
				ScoreSource:     scoreSource(contestTracking.Tracking.ScoreProvenance),
				Now:             req.Now(),
			}); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("could not update contest log %s: %w", contestTracking.ContestID, err)
			}
		}
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
