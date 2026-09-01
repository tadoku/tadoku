package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type AccountDeletionEligibilityRepository interface {
	FindRunningOwnedContestAvailableAfter(context.Context, uuid.UUID, time.Time) (*time.Time, error)
}

type AccountDeletionEligibilityRequest struct {
	UserID uuid.UUID

	checkedAt time.Time
}

func (r *AccountDeletionEligibilityRequest) CheckedAt() time.Time {
	return r.checkedAt
}

type AccountDeletionEligibilityResult struct {
	AvailableAfter *time.Time
}

func (r *AccountDeletionEligibilityResult) Eligible() bool {
	return r.AvailableAfter == nil
}

type AccountDeletionEligibility struct {
	repo  AccountDeletionEligibilityRepository
	clock commondomain.Clock
}

func NewAccountDeletionEligibility(repo AccountDeletionEligibilityRepository, clock commondomain.Clock) *AccountDeletionEligibility {
	return &AccountDeletionEligibility{repo: repo, clock: clock}
}

func (s *AccountDeletionEligibility) Execute(ctx context.Context, req *AccountDeletionEligibilityRequest) (*AccountDeletionEligibilityResult, error) {
	service := commondomain.ParseServiceIdentity(ctx)
	if service == nil {
		return nil, ErrUnauthorized
	}
	if service.Name != accountDeletionCaller {
		return nil, ErrForbidden
	}
	if req == nil || req.UserID == uuid.Nil {
		return nil, ErrRequestInvalid
	}

	req.checkedAt = s.clock.Now().UTC()
	availableAfter, err := s.repo.FindRunningOwnedContestAvailableAfter(ctx, req.UserID, req.checkedAt)
	if err != nil {
		return nil, fmt.Errorf("could not check account deletion eligibility: %w", err)
	}

	return &AccountDeletionEligibilityResult{AvailableAfter: availableAfter}, nil
}
