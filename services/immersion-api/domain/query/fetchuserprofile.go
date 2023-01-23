package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *ServiceImpl) FetchUserProfile(ctx context.Context, id uuid.UUID) (*UserProfile, error) {
	traits, err := s.kratos.FetchIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user profile: %w", err)
	}

	return &UserProfile{
		DisplayName: traits.UserDisplayName,
		CreatedAt:   traits.CreatedAt,
	}, nil
}
