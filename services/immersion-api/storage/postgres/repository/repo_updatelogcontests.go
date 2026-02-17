package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) UpdateLogContests(ctx context.Context, req *domain.LogContestUpdateDBRequest) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

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
			RegistrationID: attach.RegistrationID,
			LogID:          req.LogID,
			Amount:         req.Amount,
			Modifier:       req.Modifier,
		}); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not attach contest %s: %w", attach.ContestID, err)
		}
	}

	if err := qtx.UpdateLogEligibleOfficialLeaderboard(ctx, req.LogID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not update eligible flag: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}
