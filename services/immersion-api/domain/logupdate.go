package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogUpdateRepository interface {
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
	UpdateLog(context.Context, *LogUpdateRequest) error
	FindActivityByID(context.Context, int32) (*Activity, error)
}

type LogUpdateRequest struct {
	LogID           uuid.UUID
	UnitID          *uuid.UUID
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
	Description     *string

	// Set by domain layer (unexported: only domain can write, others read via getters)
	now      time.Time
	userID   uuid.UUID
	activity *Activity
}

func (r *LogUpdateRequest) Now() time.Time      { return r.now }
func (r *LogUpdateRequest) UserID() uuid.UUID   { return r.userID }
func (r *LogUpdateRequest) Activity() *Activity { return r.activity }

type LogUpdate struct {
	repo  LogUpdateRepository
	clock commondomain.Clock
}

func NewLogUpdate(
	repo LogUpdateRepository,
	clock commondomain.Clock,
) *LogUpdate {
	return &LogUpdate{
		repo:  repo,
		clock: clock,
	}
}

func (s *LogUpdate) Execute(ctx context.Context, req *LogUpdateRequest) (*Log, error) {
	if err := requireAuthentication(ctx); err != nil {
		return nil, err
	}

	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userID := uuid.MustParse(session.Subject)

	log, err := s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not find log to update: %w", err)
	}

	isOwner := log.UserID == userID
	if !isOwner && !isAdmin(ctx) {
		return nil, ErrForbidden
	}

	// Look up activity to determine input type
	activity, err := s.repo.FindActivityByID(ctx, int32(log.ActivityID))
	if err != nil {
		return nil, fmt.Errorf("could not find activity: %w", ErrInvalidLog)
	}

	// Validate tracking data based on input type
	if err := validateTrackingData(activity, req.DurationSeconds, req.Amount, req.UnitID); err != nil {
		return nil, err
	}

	// Store activity for repo layer to use when computing score
	req.activity = activity

	req.Tags, err = ValidateAndNormalizeTags(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("unable to validate tags: %w", err)
	}

	req.now = s.clock.Now()
	req.userID = log.UserID

	if err := s.repo.UpdateLog(ctx, req); err != nil {
		return nil, fmt.Errorf("could not update log: %w", err)
	}

	updated, err := s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not fetch updated log: %w", err)
	}

	return updated, nil
}
