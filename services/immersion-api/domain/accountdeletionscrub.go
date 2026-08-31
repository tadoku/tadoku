package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type AccountDeletionScrubRepository interface {
	ScrubAccount(context.Context, *AccountDeletionScrubRequest) error
}

type AccountDeletionScrubRequest struct {
	UserID    uuid.UUID
	RequestID uuid.UUID

	deletedAt time.Time
}

func (r *AccountDeletionScrubRequest) DeletedAt() time.Time {
	return r.deletedAt
}

type AccountDeletionScrub struct {
	repo  AccountDeletionScrubRepository
	clock commondomain.Clock
}

func NewAccountDeletionScrub(repo AccountDeletionScrubRepository, clock commondomain.Clock) *AccountDeletionScrub {
	return &AccountDeletionScrub{repo: repo, clock: clock}
}

func (s *AccountDeletionScrub) Execute(ctx context.Context, req *AccountDeletionScrubRequest) error {
	service := commondomain.ParseServiceIdentity(ctx)
	if service == nil {
		return ErrUnauthorized
	}
	if service.Name != accountDeletionCaller {
		return ErrForbidden
	}
	if req == nil || req.UserID == uuid.Nil || req.RequestID == uuid.Nil {
		return ErrRequestInvalid
	}

	req.deletedAt = s.clock.Now().UTC()
	if err := s.repo.ScrubAccount(ctx, req); err != nil {
		return fmt.Errorf("could not scrub account: %w", err)
	}

	return nil
}
