package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

// TODO: Rename to FindContestByID
func (r *Repository) FindByID(ctx context.Context, req *query.FindByIDRequest) (*query.ContestView, error) {
	contest, err := r.q.FindContestById(ctx, postgres.FindContestByIdParams{
		ID:             req.ID,
		IncludeDeleted: req.IncludeDeleted,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	activities, err := r.q.ListActivitiesForContest(ctx, contest.ID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest: %w", err)
	}

	acts := make([]query.Activity, len(activities))
	for i, a := range activities {
		acts[i] = query.Activity{
			ID:   a.ID,
			Name: a.Name,
		}
	}

	langs := []query.Language{}

	if len(contest.LanguageCodeAllowList) > 0 {
		languages, err := r.q.ListLanguagesForContest(ctx, contest.ID)
		if err != nil {
			return nil, fmt.Errorf("could not fetch contest: %w", err)
		}

		langs = make([]query.Language, len(languages))
		for i, a := range languages {
			langs[i] = query.Language{
				Code: a.Code,
				Name: a.Name,
			}
		}
	}

	return &query.ContestView{
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
