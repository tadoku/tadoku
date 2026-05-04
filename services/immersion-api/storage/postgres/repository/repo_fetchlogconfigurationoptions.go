package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FetchLogConfigurationOptions(ctx context.Context, userID uuid.UUID) (*domain.LogConfigurationOptionsResponse, error) {
	langs, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	acts, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	units, err := r.q.ListUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	userLangs, err := r.q.ListDistinctLanguageCodesForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user language codes: %w", err)
	}

	options := domain.LogConfigurationOptionsResponse{
		Languages:         make([]domain.Language, len(langs)),
		Activities:        make([]domain.Activity, len(acts)),
		Units:             make([]domain.Unit, len(units)),
		UserLanguageCodes: userLangs,
	}

	for i, l := range langs {
		options.Languages[i] = domain.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	for i, a := range acts {
		options.Activities[i] = domain.Activity{
			ID:        a.ID,
			Name:      a.Name,
			Default:   a.Default,
			InputType: a.InputType,
		}
	}

	for i, u := range units {
		options.Units[i] = domain.Unit{
			ID:            u.ID,
			LogActivityID: int(u.LogActivityID),
			Name:          u.Name,
			Modifier:      u.Modifier,
			LanguageCode:  postgres.NewStringFromNullString(u.LanguageCode),
		}
	}

	return &options, nil
}
