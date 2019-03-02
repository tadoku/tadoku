package services

import (
	"net/http"

	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// ContestService is responsible for managing contests
type ContestService interface {
	Create(ctx Context) error
}

// NewContestService initializer
func NewContestService(contestInteractor usecases.ContestInteractor) ContestService {
	return &contestService{
		ContestInteractor: contestInteractor,
	}
}

type contestService struct {
	ContestInteractor usecases.ContestInteractor
}

func (s *contestService) Create(ctx Context) error {
	contest := &domain.Contest{}
	if err := ctx.Bind(contest); err != nil {
		return fail.Wrap(err)
	}

	if err := s.ContestInteractor.CreateContest(*contest); err != nil {
		return fail.Wrap(err)
	}

	return ctx.NoContent(http.StatusCreated)
}
