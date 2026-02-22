package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func (r *Repository) FindActivityByID(ctx context.Context, id int32) (*domain.Activity, error) {
	activity, err := r.q.FindActivityByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("could not find activity: %w", err)
	}

	return &domain.Activity{
		ID:        activity.ID,
		Name:      activity.Name,
		Default:   activity.Default,
		InputType: activity.InputType,
	}, nil
}
