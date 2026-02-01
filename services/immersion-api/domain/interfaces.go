package domain

import (
	"context"

	"github.com/google/uuid"
)

// KratosClient provides identity management operations
type KratosClient interface {
	FetchIdentity(ctx context.Context, id uuid.UUID) (*UserTraits, error)
	ListIdentities(ctx context.Context, perPage int64, page int64) (*ListIdentitiesResult, error)
}

// UserCache provides cached user data
type UserCache interface {
	GetUsers() []UserCacheEntry
}

// ListIdentitiesResult contains paginated identity results
type ListIdentitiesResult struct {
	Identities []IdentityInfo
	HasMore    bool
}

// IdentityInfo contains basic identity information
type IdentityInfo struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
}
