package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FindLogByID(ctx context.Context, req *domain.LogFindRequest) (*domain.Log, error) {
	log, err := r.q.FindLogByID(ctx, postgres.FindLogByIDParams{
		IncludeDeleted: req.IncludeDeleted,
		ID:             req.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch log details: %w", err)
	}

	registrations, err := r.q.FindAttachedContestRegistrationsForLog(ctx, req.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("could not fetch log details: %w", err)
	}

	refs := make([]domain.ContestRegistrationReference, len(registrations))
	for i, it := range registrations {
		refs[i] = domain.ContestRegistrationReference{
			RegistrationID:       it.ID,
			ContestID:            it.ContestID,
			ContestEnd:           it.ContestEnd,
			Title:                it.Title,
			OwnerUserDisplayName: it.OwnerUserDisplayName,
			Official:             it.Official,
			Score:                float32(it.Score.Float64),
		}
	}

	// UnitName from FindLogByIDRow is string (not nullable per sqlc), but with
	// left join it could be empty string for time-only logs. Treat empty as nil.
	var unitName *string
	if log.UnitName != "" {
		unitName = &log.UnitName
	}

	return &domain.Log{
		ID:                          log.ID,
		UserID:                      log.UserID,
		UserDisplayName:             &log.UserDisplayName,
		Description:                 postgres.NewStringFromNullString(log.Description),
		LanguageCode:                log.LanguageCode,
		LanguageName:                log.LanguageName,
		ActivityID:                  int(log.ActivityID),
		ActivityName:                log.ActivityName,
		UnitID:                      postgres.NewUUIDFromNullUUID(log.UnitID),
		UnitName:                    unitName,
		Tags:                        postgres.StringArrayFromInterface(log.Tags),
		Amount:                      postgres.NewFloat32FromNullFloat64(log.Amount),
		Modifier:                    postgres.NewFloat32FromNullFloat64(log.Modifier),
		Score:                       log.Score,
		DurationSeconds:             postgres.NewInt32FromNullInt32(log.DurationSeconds),
		EligibleOfficialLeaderboard: log.EligibleOfficialLeaderboard,
		CreatedAt:                   log.CreatedAt,
		UpdatedAt:                   log.UpdatedAt,
		Deleted:                     log.DeletedAt.Valid,
		Registrations:               refs,
	}, nil
}
