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

	if commondomain.IsRole(ctx, commondomain.RoleGuest) || commondomain.IsRole(ctx, commondomain.RoleBanned) || session == nil {
		return nil, ErrUnauthorized
	}

	req.UserID = uuid.MustParse(session.Subject)

	return s.repo.FindRegistrationForUser(ctx, req)
}
