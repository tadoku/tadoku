package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

type Repository interface {
	// contest
	CreateContest(context.Context, *domain.ContestCreateRequest) (*domain.ContestCreateResponse, error)
	UpsertContestRegistration(context.Context, *domain.RegistrationUpsertRequest) error
	FetchOngoingContestRegistrations(context.Context, *domain.RegistrationListOngoingRequest) (*domain.ContestRegistrations, error)

	// log
	CreateLog(context.Context, *domain.LogCreateRequest) (*uuid.UUID, error)
	DeleteLog(context.Context, *domain.LogDeleteRequest) error
	DetachLogFromContest(context.Context, *domain.ContestModerationDetachLogRequest, uuid.UUID) error

	UpsertUser(context.Context, *domain.UserUpsertRequest) error

	query.Repository
}
