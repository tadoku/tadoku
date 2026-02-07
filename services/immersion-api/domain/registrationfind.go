package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type RegistrationFindRepository interface {
	FindRegistrationForUser(context.Context, *RegistrationFindRequest) (*ContestRegistration, error)
}

type RegistrationFindRequest struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

type RegistrationFind struct {
	repo RegistrationFindRepository
}

func NewRegistrationFind(repo RegistrationFindRepository) *RegistrationFind {
	return &RegistrationFind{repo: repo}
}

func (s *RegistrationFind) Execute(ctx context.Context, req *RegistrationFindRequest) (*ContestRegistration, error) {
	session := commondomain.ParseUserIdentity(ctx)

	if isGuest(ctx) || session == nil {
		return nil, ErrUnauthorized
	}
	if err := requireNotBannedUnauthorized(ctx); err != nil {
		return nil, err
	}

	req.UserID = uuid.MustParse(session.Subject)

	return s.repo.FindRegistrationForUser(ctx, req)
}
