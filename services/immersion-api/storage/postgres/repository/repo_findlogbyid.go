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

	tracking := readLogTracking(log.UnitID, log.Amount, log.Modifier, log.DurationSeconds, log.Score)

	return &domain.Log{
		ID:                          log.ID,
		UserID:                      log.UserID,
		UserDisplayName:             &log.UserDisplayName,
		Description:                 postgres.NewStringFromNullString(log.Description),
		LanguageCode:                log.LanguageCode,
		LanguageName:                log.LanguageName,
		ActivityID:                  int(log.ActivityID),
		UnitID:                      postgres.NewUUIDFromNullUUID(log.UnitID),
		UnitName:                    log.UnitName,
		Tags:                        postgres.StringArrayFromInterface(log.Tags),
		Amount:                      postgres.NewFloat32FromNullFloat64(log.Amount),
		Modifier:                    postgres.NewFloat32FromNullFloat64(log.Modifier),
		Score:                       postgres.NewFloat32FromNullFloat64(log.Score),
		DurationSeconds:             postgres.NewInt32PtrFromNullInt32(log.DurationSeconds),
		Tracking:                    tracking,
		EligibleOfficialLeaderboard: log.EligibleOfficialLeaderboard,
		CreatedAt:                   log.CreatedAt,
		UpdatedAt:                   log.UpdatedAt,
		Deleted:                     log.DeletedAt.Valid,
		Registrations:               refs,
	}, nil
}
