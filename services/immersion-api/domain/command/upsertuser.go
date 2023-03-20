package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type UpsertUserRequest struct {
	ID               uuid.UUID
	DisplayName      string
	SessionCreatedAt time.Time
}

func (s *ServiceImpl) UpdateUserMetadataFromSession(ctx context.Context) error {
	// Enrich request with session
	session := domain.ParseSession(ctx)
	if session == nil {
		return ErrUnauthorized
	}

	req := &UpsertUserRequest{
		ID:               uuid.MustParse(session.Subject),
		DisplayName:      session.DisplayName,
		SessionCreatedAt: session.CreatedAt,
	}

	return s.r.UpsertUser(ctx, req)
}
