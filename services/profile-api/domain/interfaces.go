package domain

import (
	"context"
)

// KratosClient provides identity management operations
type KratosClient interface {
	ListIdentities(ctx context.Context, perPage int64, page int64) (*ListIdentitiesResult, error)
}

// UserListCache provides cached user data
type UserListCache interface {
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
