package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogDeleteRepository interface {
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
	DeleteLog(context.Context, *LogDeleteRequest) error
}

type LogDeleteRequest struct {
	LogID uuid.UUID

	// Set by domain layer (unexported: only domain can write, others read via getters)
	now time.Time
}

func (r *LogDeleteRequest) Now() time.Time { return r.now }

type LogDelete struct {
	repo  LogDeleteRepository
	clock commondomain.Clock
}

func NewLogDelete(
	repo LogDeleteRepository,
	clock commondomain.Clock,
) *LogDelete {
	return &LogDelete{
		repo:  repo,
		clock: clock,
	}
}

func (s *LogDelete) Execute(ctx context.Context, req *LogDeleteRequest) error {
	if err := requireAuthentication(ctx); err != nil {
		return err
	}

	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return ErrUnauthorized
	}

	log, err := s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return fmt.Errorf("could not find log to delete: %w", err)
	}

	isOwner := log.UserID == uuid.MustParse(session.Subject)
	if !isOwner && !isAdmin(ctx) {
		return ErrForbidden
	}

	req.now = s.clock.Now()

	if err := s.repo.DeleteLog(ctx, req); err != nil {
		return err
	}

	return nil
}
