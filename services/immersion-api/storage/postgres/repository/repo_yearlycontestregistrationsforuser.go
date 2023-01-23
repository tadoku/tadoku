package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) YearlyContestRegistrationsForUser(ctx context.Context, req *query.YearlyContestRegistrationsForUserRequest) (*query.ContestRegistrations, error) {
	regs, err := r.q.FindYearlyContestRegistrationForUser(ctx, postgres.FindYearlyContestRegistrationForUserParams{
		UserID:         req.UserID,
		Year:           int32(req.Year),
		IncludePrivate: req.IncludePrivate,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.ContestRegistrations{
				Registrations: []query.ContestRegistration{},
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}
		return nil, fmt.Errorf("could not fetch contest registrations: %w", err)
	}

	languages, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest registrations: %w", err)
	}

	activities, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest registrations: %w", err)
	}

	langs := map[string]string{}
	acts := map[int32]string{}

	for _, l := range languages {
		langs[l.Code] = l.Name
	}
	for _, a := range activities {
		acts[a.ID] = a.Name
	}

	res := &query.ContestRegistrations{
		Registrations: make([]query.ContestRegistration, len(regs)),
		TotalSize:     len(regs),
		NextPageToken: "",
	}
	for i, r := range regs {
		r := r

		// TODO: refactor this out to a mapper
		contest := &query.ContestView{
			ID:                r.ContestID,
			ContestStart:      r.ContestStart,
			ContestEnd:        r.ContestEnd,
			RegistrationEnd:   r.RegistrationEnd,
			Title:             r.Title,
			Description:       postgres.NewStringFromNullString(r.Description),
			Private:           r.Private,
			Official:          r.Official,
			AllowedLanguages:  make([]query.Language, 0),
			AllowedActivities: make([]query.Activity, len(r.ActivityTypeIDAllowList)),
		}

		for i, a := range r.ActivityTypeIDAllowList {
			contest.AllowedActivities[i] = query.Activity{
				ID:   a,
				Name: acts[a],
			}
		}

		reg := query.ContestRegistration{
			ID:              r.ID,
			ContestID:       r.ContestID,
			UserID:          r.UserID,
			UserDisplayName: r.UserDisplayName,
			Languages:       make([]query.Language, len(r.LanguageCodes)),
			Contest:         contest,
		}

		for i, code := range r.LanguageCodes {
			reg.Languages[i] = query.Language{
				Code: code,
				Name: langs[code],
			}
		}

		res.Registrations[i] = reg
	}

	return res, nil
}
