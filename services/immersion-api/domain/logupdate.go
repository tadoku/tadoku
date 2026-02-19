package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogUpdateRepository interface {
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
	UpdateLog(context.Context, *LogUpdateRequest) error
}

type LogUpdateRequest struct {
	LogID       uuid.UUID
	UnitID      uuid.UUID `validate:"required"`
	Amount      float32   `validate:"required,gte=0"`
	Tags        []string
	Description *string

	// Set by domain layer (unexported: only domain can write, others read via getters)
	now    time.Time
	userID uuid.UUID
}

func (r *LogUpdateRequest) Now() time.Time    { return r.now }
func (r *LogUpdateRequest) UserID() uuid.UUID { return r.userID }

type LogUpdate struct {
	repo     LogUpdateRepository
	clock    commondomain.Clock
	validate *validator.Validate
}

func NewLogUpdate(
	repo LogUpdateRepository,
	clock commondomain.Clock,
) *LogUpdate {
	return &LogUpdate{
		repo:     repo,
		clock:    clock,
		validate: validator.New(),
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

	err = s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("unable to validate: %w", ErrInvalidLog)
	}

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
