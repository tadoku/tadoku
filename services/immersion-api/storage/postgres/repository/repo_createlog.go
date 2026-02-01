package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) CreateLog(ctx context.Context, req *domain.LogCreateRequest) (*uuid.UUID, error) {
	unit, err := r.q.FindUnitForTracking(ctx, postgres.FindUnitForTrackingParams{
		ID:            req.UnitID,
		LogActivityID: int16(req.ActivityID),
		LanguageCode:  postgres.NewNullString(&req.LanguageCode),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid unit supplied: %w", domain.ErrInvalidLog)
		}
		return nil, fmt.Errorf("could not fetch unit for tracking: %w", err)
	}

	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create log: %w", err)
	}
	qtx := r.q.WithTx(tx)

	id := uuid.New()
	logId, err := qtx.CreateLog(ctx, postgres.CreateLogParams{
		ID:                          id,
		UserID:                      req.UserID,
		LanguageCode:                req.LanguageCode,
		LogActivityID:               int16(req.ActivityID),
		UnitID:                      req.UnitID,
		Tags:                        req.Tags,
		Amount:                      req.Amount,
		Modifier:                    unit.Modifier,
		EligibleOfficialLeaderboard: req.EligibleOfficialLeaderboard,
		Description:                 postgres.NewNullString(req.Description),
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create log: %w", err)
	}

	for _, registrationID := range req.RegistrationIDs {
		if err = qtx.CreateContestLogRelation(ctx, postgres.CreateContestLogRelationParams{
			RegistrationID: registrationID,
			LogID:          id,
		}); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("could not create log: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not create log: %w", err)
	}

	return &logId, nil
}
