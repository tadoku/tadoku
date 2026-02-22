package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) UpdateLogContests(ctx context.Context, req *domain.LogContestUpdateDBRequest) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	// Fetch log context before changes to capture pre-change eligible flag
	logCtxBefore, err := qtx.FetchLogOutboxContext(ctx, req.LogID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not fetch log context: %w", err)
	}

	for _, contestID := range req.Detach {
		if err := qtx.DetachLogFromContest(ctx, postgres.DetachLogFromContestParams{
			ContestID: contestID,
			LogID:     req.LogID,
		}); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not detach contest %s: %w", contestID, err)
		}
	}

	for _, attach := range req.Attach {
		if err := qtx.CreateContestLogRelation(ctx, postgres.CreateContestLogRelationParams{
			RegistrationID:  attach.RegistrationID,
			LogID:           req.LogID,
			Amount:          postgres.NewNullFloat64(req.Amount),
			Modifier:        postgres.NewNullFloat64(req.Modifier),
			DurationSeconds: postgres.NewNullInt32(req.DurationSeconds),
			ComputedScore:   postgres.NewNullFloat64(req.ComputedScore),
		}); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not attach contest %s: %w", attach.ContestID, err)
		}
	}

	if err := qtx.UpdateLogEligibleOfficialLeaderboard(ctx, req.LogID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not update eligible flag: %w", err)
	}

	// Fetch log context after changes to capture post-change eligible flag
	logCtxAfter, err := qtx.FetchLogOutboxContext(ctx, req.LogID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not fetch updated log context: %w", err)
	}

	// Collect all affected contest IDs (both attached and detached)
	contestIDSet := map[uuid.UUID]struct{}{}
	for _, contestID := range req.Detach {
		contestIDSet[contestID] = struct{}{}
	}
	for _, attach := range req.Attach {
		contestIDSet[attach.ContestID] = struct{}{}
	}
	allContestIDs := make([]uuid.UUID, 0, len(contestIDSet))
	for id := range contestIDSet {
		allContestIDs = append(allContestIDs, id)
	}

	// Emit outbox events — refresh official if either before or after is eligible
	if err = insertLeaderboardOutboxEvents(ctx, qtx, LeaderboardOutboxParams{
		UserID:          logCtxBefore.UserID,
		ContestIDs:      allContestIDs,
		OfficialContest: logCtxBefore.EligibleOfficialLeaderboard || logCtxAfter.EligibleOfficialLeaderboard,
		Year:            logCtxBefore.Year,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}
