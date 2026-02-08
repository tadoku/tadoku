package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type RegistrationListYearlyRepository interface {
	YearlyContestRegistrationsForUser(context.Context, *RegistrationListYearlyRequest) (*ContestRegistrations, error)
}

type RegistrationListYearlyRequest struct {
	UserID         uuid.UUID
	Year           int
	IncludePrivate bool
}

type RegistrationListYearly struct {
	repo RegistrationListYearlyRepository
}

func NewRegistrationListYearly(repo RegistrationListYearlyRepository) *RegistrationListYearly {
	return &RegistrationListYearly{repo: repo}
}

func (s *RegistrationListYearly) Execute(ctx context.Context, req *RegistrationListYearlyRequest) (*ContestRegistrations, error) {
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(session.Subject)

	sessionMatchesUser := err == nil && userID == req.UserID
	req.IncludePrivate = isAdmin(ctx) || sessionMatchesUser

	return s.repo.YearlyContestRegistrationsForUser(ctx, req)
}
