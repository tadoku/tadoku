package command

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

var ErrInvalidLog = errors.New("unable to validate log")
var ErrInvalidContest = errors.New("unable to validate contest")
var ErrInvalidContestRegistration = errors.New("language selection is not valid for contest")
var ErrRequestInvalid = errors.New("request is invalid")
var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("not allowed")
var ErrUnauthorized = errors.New("unauthorized")

type Repository interface {
	// contest
	CreateContest(context.Context, *CreateContestRequest) (*CreateContestResponse, error)
	UpsertContestRegistration(context.Context, *UpsertContestRegistrationRequest) error

	// log
	CreateLog(context.Context, *CreateLogRequest) (*uuid.UUID, error)
	DeleteLog(context.Context, *DeleteLogRequest) error
	DetachLogFromContest(context.Context, *DetachLogFromContestRequest, uuid.UUID) error

	UpsertUser(context.Context, *UpsertUserRequest) error

	query.Repository
}

type Service interface {
	// contest
	CreateContest(context.Context, *CreateContestRequest) (*CreateContestResponse, error)
	UpsertContestRegistration(context.Context, *UpsertContestRegistrationRequest) error

	// log
	CreateLog(context.Context, *CreateLogRequest) (*query.Log, error)
	DeleteLog(context.Context, *DeleteLogRequest) error
	DetachLogFromContest(context.Context, *DetachLogFromContestRequest) error

	UpdateUserMetadataFromSession(context.Context) error
}

type ServiceImpl struct {
	r        Repository
	validate *validator.Validate
	clock    domain.Clock
}

func NewService(r Repository, clock domain.Clock) Service {
	return &ServiceImpl{
		r:        r,
		validate: validator.New(),
		clock:    clock,
	}
}
