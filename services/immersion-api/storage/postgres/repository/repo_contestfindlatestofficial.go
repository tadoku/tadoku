package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) ContestFindLatestOfficial(ctx context.Context) (*domain.ContestView, error) {
	contest, err := r.q.FindLatestOfficialContest(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	activities, err := r.q.ListActivitiesForContest(ctx, contest.ID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	acts := make([]domain.Activity, len(activities))
	for i, a := range activities {
		acts[i] = domain.Activity{
			ID:           a.ID,
			Name:         a.Name,
			TimeModifier: a.TimeModifier,
			InputType:    a.InputType,
		}
	}

	langs := []domain.Language{}

	if len(contest.LanguageCodeAllowList) > 0 {
		languages, err := r.q.ListLanguagesForContest(ctx, contest.ID)
		if err != nil {
			return nil, fmt.Errorf("could not fetch contest: %w", err)
		}

		langs = make([]domain.Language, len(languages))
		for i, a := range languages {
			langs[i] = domain.Language{
				Code: a.Code,
				Name: a.Name,
			}
		}
	}

	return &domain.ContestView{
		ID:                   contest.ID,
		ContestStart:         contest.ContestStart,
		ContestEnd:           contest.ContestEnd,
		RegistrationEnd:      contest.RegistrationEnd,
		Title:                contest.Title,
		Description:          postgres.NewStringFromNullString(contest.Description),
		OwnerUserID:          contest.OwnerUserID,
		OwnerUserDisplayName: contest.OwnerUserDisplayName,
		Official:             contest.Official,
		Private:              contest.Private,
		AllowedLanguages:     langs,
		AllowedActivities:    acts,
		CreatedAt:            contest.CreatedAt,
		UpdatedAt:            contest.UpdatedAt,
		Deleted:              contest.DeletedAt.Valid,
	}, nil
}
