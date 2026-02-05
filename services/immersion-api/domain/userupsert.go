package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type UserUpsertRepository interface {
	UpsertUser(context.Context, *UserUpsertRequest) error
}

type UserUpsertRequest struct {
	ID               uuid.UUID
	DisplayName      string
	SessionCreatedAt time.Time
}

type UserUpsert struct {
	repo UserUpsertRepository
}

func NewUserUpsert(repo UserUpsertRepository) *UserUpsert {
	return &UserUpsert{repo: repo}
}

func (s *UserUpsert) Execute(ctx context.Context) error {
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return ErrUnauthorized
	}

	req := &UserUpsertRequest{
		ID:               uuid.MustParse(session.Subject),
		DisplayName:      session.DisplayName,
		SessionCreatedAt: session.CreatedAt,
	}

	return s.repo.UpsertUser(ctx, req)
}
