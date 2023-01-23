package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FindRegistrationForUser(ctx context.Context, req *query.FindRegistrationForUserRequest) (*query.ContestRegistration, error) {
	reg, err := r.q.FindContestRegistrationForUser(ctx, postgres.FindContestRegistrationForUserParams{
		UserID:    req.UserID,
		ContestID: req.ContestID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch contest registration: %w", err)
	}

	languages, err := r.q.GetLanguagesByCode(ctx, reg.LanguageCodes)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest registrations: %w", err)
	}

	registrationLanguages := make([]query.Language, len(reg.LanguageCodes))
	for i, it := range languages {
		registrationLanguages[i] = query.Language{
			Code: it.Code,
			Name: it.Name,
		}
	}

	contest := &query.ContestView{
		ID:                reg.ContestID,
		ContestStart:      reg.ContestStart,
		ContestEnd:        reg.ContestEnd,
		RegistrationEnd:   reg.RegistrationEnd,
		Title:             reg.Title,
		Description:       postgres.NewStringFromNullString(reg.Description),
		Private:           reg.Private,
		Official:          reg.Official,
		AllowedLanguages:  make([]query.Language, 0),
		AllowedActivities: make([]query.Activity, 0),
	}

	return &query.ContestRegistration{
		ID:              reg.ID,
		ContestID:       reg.ContestID,
		UserID:          req.UserID,
		UserDisplayName: reg.UserDisplayName,
		Languages:       registrationLanguages,
		Contest:         contest,
	}, nil
}
