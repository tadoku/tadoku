package domain

import (
	"context"

	"github.com/google/uuid"
)

// KratosClient provides identity management operations
type KratosClient interface {
	FetchIdentity(ctx context.Context, id uuid.UUID) (*UserTraits, error)
}
