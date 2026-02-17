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
		Amount:                      req.Amount,
		Modifier:                    unit.Modifier,
		EligibleOfficialLeaderboard: req.EligibleOfficialLeaderboard,
		Description:                 postgres.NewNullString(req.Description),
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create log: %w", err)
	}

	// Track unique contest IDs for outbox events
	contestIDSet := map[uuid.UUID]struct{}{}

	for _, registrationID := range req.RegistrationIDs {
		if err = qtx.CreateContestLogRelation(ctx, postgres.CreateContestLogRelationParams{
			RegistrationID: registrationID,
			LogID:          id,
			Amount:         req.Amount,
			Modifier:       unit.Modifier,
		}); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("could not create log: %w", err)
		}

		contestID, err := qtx.FetchContestIDForRegistration(ctx, registrationID)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("could not resolve contest for registration: %w", err)
		}
		contestIDSet[contestID] = struct{}{}
	}

	// Insert tags into log_tags table
	for _, tag := range req.Tags {
		if err = qtx.InsertLogTag(ctx, postgres.InsertLogTagParams{
			LogID:  id,
			UserID: req.UserID,
			Tag:    tag,
		}); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("could not insert log tag: %w", err)
		}
	}

	// Write outbox events for leaderboard sync
	contestIDs := make([]uuid.UUID, 0, len(contestIDSet))
	for id := range contestIDSet {
		contestIDs = append(contestIDs, id)
	}
	if err = insertLeaderboardOutboxEvents(ctx, qtx, LeaderboardOutboxParams{
		UserID:          req.UserID,
		ContestIDs:      contestIDs,
		OfficialContest: req.EligibleOfficialLeaderboard,
		Year:            req.Year,
	}); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not create log: %w", err)
	}

	return &logId, nil
}
