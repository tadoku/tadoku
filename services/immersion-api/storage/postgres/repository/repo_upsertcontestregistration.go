package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) UpsertContestRegistration(ctx context.Context, req *domain.RegistrationUpsertRequest) error {
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
