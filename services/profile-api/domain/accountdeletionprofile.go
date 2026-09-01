package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type AccountDeletionProfileRepository interface {
	DeleteProfile(context.Context, uuid.UUID) error
}

type AccountDeletionProfileCache interface {
	SuppressAndEvict(uuid.UUID)
}

type AccountDeletionProfile struct {
	repo      AccountDeletionProfileRepository
	userCache AccountDeletionProfileCache
}

func NewAccountDeletionProfile(repo AccountDeletionProfileRepository, userCache AccountDeletionProfileCache) *AccountDeletionProfile {
	return &AccountDeletionProfile{repo: repo, userCache: userCache}
}

func (s *AccountDeletionProfile) Execute(ctx context.Context, identityID uuid.UUID) error {
	if identityID == uuid.Nil {
		return ErrRequestInvalid
	}

	if err := s.repo.DeleteProfile(ctx, identityID); err != nil {
		return fmt.Errorf("could not delete profile: %w", err)
	}
	s.userCache.SuppressAndEvict(identityID)
	return nil
}
