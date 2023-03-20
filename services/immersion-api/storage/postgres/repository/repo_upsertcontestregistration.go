package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) UpsertContestRegistration(ctx context.Context, req *command.UpsertContestRegistrationRequest) error {
	_, err := r.q.UpsertContestRegistration(ctx, postgres.UpsertContestRegistrationParams{
		ID:            req.ID,
		ContestID:     req.ContestID,
		UserID:        req.UserID,
		LanguageCodes: req.LanguageCodes,
	})

	if err != nil {
		return fmt.Errorf("could not create or update contest registration: %w", err)
	}

	return nil
}
