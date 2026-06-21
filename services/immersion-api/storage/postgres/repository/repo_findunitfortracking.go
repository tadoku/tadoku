package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FindUnitForTracking(ctx context.Context, req *domain.UnitFindForTrackingRequest) (*domain.Unit, error) {
	unit, err := r.q.FindUnitForTracking(ctx, postgres.FindUnitForTrackingParams{
		ID:            req.ID,
		LogActivityID: int16(req.ActivityID),
		LanguageCode:  postgres.NewNullString(&req.LanguageCode),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid unit supplied: %w", domain.ErrInvalidLog)
		}
		return nil, fmt.Errorf("could not fetch unit for tracking: %w", err)
	}

	return &domain.Unit{
		ID:            unit.ID,
		LogActivityID: int(unit.LogActivityID),
		Name:          unit.Name,
		Modifier:      unit.Modifier,
		LanguageCode:  postgres.NewStringFromNullString(unit.LanguageCode),
	}, nil
}
