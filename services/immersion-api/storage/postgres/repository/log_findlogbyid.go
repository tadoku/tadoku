package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FindLogByID(ctx context.Context, req *query.FindLogByIDRequest) (*query.Log, error) {
	log, err := r.q.FindLogByID(ctx, postgres.FindLogByIDParams{
		IncludeDeleted: req.IncludeDeleted,
		ID:             req.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch log details: %w", err)
	}

	registrations, err := r.q.FindAttachedContestRegistrationsForLog(ctx, req.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("could not fetch log details: %w", err)
	}

	refs := make([]query.ContestRegistrationReference, len(registrations))
	for i, it := range registrations {
		refs[i] = query.ContestRegistrationReference{
			RegistrationID: it.ID,
			ContestID:      it.ContestID,
			Title:          it.Title,
		}
	}

	return &query.Log{
		ID:              log.ID,
		UserID:          log.UserID,
		UserDisplayName: &log.UserDisplayName,
		Description:     postgres.NewStringFromNullString(log.Description),
		LanguageCode:    log.LanguageCode,
		LanguageName:    log.LanguageName,
		ActivityID:      int(log.ActivityID),
		ActivityName:    log.ActivityName,
		UnitName:        log.UnitName,
		Tags:            log.Tags,
		Amount:          log.Amount,
		Modifier:        log.Modifier,
		Score:           log.Score,
		CreatedAt:       log.CreatedAt,
		UpdatedAt:       log.UpdatedAt,
		Deleted:         log.DeletedAt.Valid,
		Registrations:   refs,
	}, nil
}
