package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) FindUserDisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	rows, err := r.q.FindUserDisplayNames(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user display names: %w", err)
	}

	names := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		names[row.ID] = row.DisplayName
	}

	return names, nil
}
