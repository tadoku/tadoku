package command

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

var ErrRequestInvalid = errors.New("request is invalid")
var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("not allowed")
var ErrUnauthorized = errors.New("unauthorized")

type Repository interface {
	// contest
	CreateContest(context.Context, *ContestCreateRequest) (*ContestCreateResponse, error)
	UpsertContestRegistration(context.Context, *UpsertContestRegistrationRequest) error

	// log
	CreateLog(context.Context, *LogCreateRequest) error

	query.Repository
}

type Service interface {
	// contest
	CreateContest(context.Context, *ContestCreateRequest) (*ContestCreateResponse, error)
	UpsertContestRegistration(context.Context, *UpsertContestRegistrationRequest) error

	// log
	CreateLog(context.Context, *LogCreateRequest) error
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
