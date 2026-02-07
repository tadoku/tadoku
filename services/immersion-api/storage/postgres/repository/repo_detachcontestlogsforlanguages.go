package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) DetachContestLogsForLanguages(ctx context.Context, req *domain.DetachContestLogsForLanguagesRequest) error {
	err := r.q.DetachContestLogsForLanguages(ctx, postgres.DetachContestLogsForLanguagesParams{
		ContestID:     req.ContestID,
		UserID:        req.UserID,
		LanguageCodes: req.LanguageCodes,
	})

	if err != nil {
		return fmt.Errorf("could not detach contest logs for languages: %w", err)
	}

	return nil
}
