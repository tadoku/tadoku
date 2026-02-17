package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) UpsertContestRegistration(ctx context.Context, req *domain.RegistrationUpsertRequest) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	_, err = qtx.UpsertContestRegistration(ctx, postgres.UpsertContestRegistrationParams{
		ID:            req.ID,
		ContestID:     req.ContestID,
		UserID:        req.UserID,
		LanguageCodes: req.LanguageCodes,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not create or update contest registration: %w", err)
	}

	// Write outbox events for leaderboard sync
	if err = insertLeaderboardOutboxEvents(ctx, qtx, LeaderboardOutboxParams{
		UserID:          req.UserID,
		ContestIDs:      []uuid.UUID{req.ContestID},
		OfficialContest: req.OfficialContest,
		Year:            req.Year,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not create or update contest registration: %w", err)
	}

	return nil
}
