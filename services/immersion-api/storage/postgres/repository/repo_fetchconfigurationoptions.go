package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func (r *Repository) FetchContestConfigurationOptions(ctx context.Context) (*domain.ContestConfigurationOptionsResponse, error) {
	langs, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest configuration options: %w", err)
	}

	acts, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest configuration options: %w", err)
	}

	options := domain.ContestConfigurationOptionsResponse{
		Languages:  make([]domain.Language, len(langs)),
		Activities: make([]domain.Activity, len(acts)),
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

	return &options, err
}
