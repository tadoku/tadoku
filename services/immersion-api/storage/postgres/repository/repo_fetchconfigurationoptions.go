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

	options := domain.ContestConfigurationOptionsResponse{
		Languages: make([]domain.Language, len(langs)),
	}

	for i, l := range langs {
		options.Languages[i] = domain.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	return &options, err
}
