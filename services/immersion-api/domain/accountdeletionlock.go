package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

const accountDeletionCaller = "profile-api"

type AccountDeletionLockRepository interface {
	LockAccountForDeletion(context.Context, *AccountDeletionLockRequest) error
}

type AccountDeletionLockRequest struct {
	UserID    uuid.UUID
	RequestID uuid.UUID

	lockedAt time.Time
}

func (r *AccountDeletionLockRequest) LockedAt() time.Time {
	return r.lockedAt
}

type AccountDeletionLock struct {
	repo  AccountDeletionLockRepository
	clock commondomain.Clock
}

func NewAccountDeletionLock(repo AccountDeletionLockRepository, clock commondomain.Clock) *AccountDeletionLock {
	return &AccountDeletionLock{repo: repo, clock: clock}
}

func (s *AccountDeletionLock) Execute(ctx context.Context, req *AccountDeletionLockRequest) error {
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

	req.lockedAt = s.clock.Now().UTC()
	if err := s.repo.LockAccountForDeletion(ctx, req); err != nil {
		return fmt.Errorf("could not lock account for deletion: %w", err)
	}

	return nil
}
