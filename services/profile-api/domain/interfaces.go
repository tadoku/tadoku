package domain

import (
	"context"
)

// KratosClient provides identity management operations
type KratosClient interface {
	ListIdentities(ctx context.Context) ([]IdentityInfo, error)
}

// UserListCache provides cached user data
type UserListCache interface {
	GetUsers() []UserCacheEntry
}

// IdentityInfo contains basic identity information
type IdentityInfo struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
}
