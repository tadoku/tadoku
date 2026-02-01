package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type RegistrationListOngoingRepository interface {
	FetchOngoingContestRegistrations(context.Context, *RegistrationListOngoingRequest) (*ContestRegistrations, error)
}

type RegistrationListOngoingRequest struct {
	UserID uuid.UUID
	Now    time.Time
}

type RegistrationListOngoing struct {
	repo  RegistrationListOngoingRepository
	clock commondomain.Clock
}

func NewRegistrationListOngoing(repo RegistrationListOngoingRepository, clock commondomain.Clock) *RegistrationListOngoing {
	return &RegistrationListOngoing{repo: repo, clock: clock}
}

func (s *RegistrationListOngoing) Execute(ctx context.Context) (*ContestRegistrations, error) {
	if commondomain.IsRole(ctx, commondomain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	session := commondomain.ParseSession(ctx)
	req := &RegistrationListOngoingRequest{
		UserID: uuid.MustParse(session.Subject),
		Now:    s.clock.Now(),
	}

	return s.repo.FetchOngoingContestRegistrations(ctx, req)
}
