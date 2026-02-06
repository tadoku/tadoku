package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func (r *Repository) ListLanguages(ctx context.Context) ([]domain.Language, error) {
	rows, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list languages: %w", err)
	}

	languages := make([]domain.Language, len(rows))
	for i, row := range rows {
		languages[i] = domain.Language{
			Code: row.Code,
			Name: row.Name,
		}
	}

	return languages, nil
}
