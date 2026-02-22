package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FetchOngoingContestRegistrations(ctx context.Context, req *domain.RegistrationListOngoingRequest) (*domain.ContestRegistrations, error) {
	regs, err := r.q.FindOngoingContestRegistrationForUser(ctx, postgres.FindOngoingContestRegistrationForUserParams{
		UserID: req.UserID,
		Now:    req.Now,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.ContestRegistrations{
				Registrations: []domain.ContestRegistration{},
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}
		return nil, fmt.Errorf("could not fetch ongoing contest registrations: %w", err)
	}

	languages, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch ongoing contest registrations: %w", err)
	}

	activities, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch ongoing contest registrations: %w", err)
	}

	langs := map[string]string{}
	type actInfo struct {
		Name      string
		InputType string
	}
	acts := map[int32]actInfo{}

	for _, l := range languages {
		langs[l.Code] = l.Name
	}
	for _, a := range activities {
		acts[a.ID] = actInfo{Name: a.Name, InputType: a.InputType}
	}

	res := &domain.ContestRegistrations{
		Registrations: make([]domain.ContestRegistration, len(regs)),
		TotalSize:     len(regs),
		NextPageToken: "",
	}
	for i, r := range regs {
		r := r

		contest := &domain.ContestView{
			ID:                   r.ContestID,
			ContestStart:         r.ContestStart,
			ContestEnd:           r.ContestEnd,
			RegistrationEnd:      r.RegistrationEnd,
			Title:                r.Title,
			Description:          postgres.NewStringFromNullString(r.Description),
			Private:              r.Private,
			Official:             r.Official,
			OwnerUserID:          r.OwnerUserID,
			OwnerUserDisplayName: r.OwnerUserDisplayName,
			AllowedLanguages:     make([]domain.Language, 0),
			AllowedActivities:    make([]domain.Activity, len(r.ActivityTypeIDAllowList)),
		}

		for i, a := range r.ActivityTypeIDAllowList {
			info := acts[a]
			contest.AllowedActivities[i] = domain.Activity{
				ID:        a,
				Name:      info.Name,
				InputType: info.InputType,
			}
		}

		reg := domain.ContestRegistration{
			ID:              r.ID,
			ContestID:       r.ContestID,
			UserID:          r.UserID,
			UserDisplayName: r.UserDisplayName,
			Languages:       make([]domain.Language, len(r.LanguageCodes)),
			Contest:         contest,
		}

		for i, code := range r.LanguageCodes {
			reg.Languages[i] = domain.Language{
				Code: code,
				Name: langs[code],
			}
		}

		res.Registrations[i] = reg
	}

	return res, nil
}
