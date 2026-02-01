package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ProfileFetchKratosClient interface {
	FetchIdentity(ctx context.Context, id uuid.UUID) (*UserTraits, error)
}

type ProfileFetchRequest struct {
	UserID uuid.UUID
}

type ProfileFetch struct {
	kratos ProfileFetchKratosClient
}

func NewProfileFetch(kratos ProfileFetchKratosClient) *ProfileFetch {
	return &ProfileFetch{kratos: kratos}
}

func (s *ProfileFetch) Execute(ctx context.Context, req *ProfileFetchRequest) (*UserProfile, error) {
	traits, err := s.kratos.FetchIdentity(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user profile: %w", err)
	}

	return &UserProfile{
		DisplayName: traits.UserDisplayName,
		CreatedAt:   traits.CreatedAt,
	}, nil
}
