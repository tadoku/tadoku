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
	Update(ctx Context) error
	Latest(ctx Context) error
	Get(ctx Context) error
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

func (s *contestService) Update(ctx Context) error {
	contest := &domain.Contest{}
	if err := ctx.Bind(contest); err != nil {
		return fail.Wrap(err)
	}

	ctx.BindID(&contest.ID)

	if err := s.ContestInteractor.UpdateContest(*contest); err != nil {
		return fail.Wrap(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (s *contestService) Latest(ctx Context) error {
	contest, err := s.ContestInteractor.Latest()

	if err != nil {
		if err == usecases.ErrContestNotFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return fail.Wrap(err)
	}

	return ctx.JSON(http.StatusOK, contest)
}

func (s *contestService) Get(ctx Context) error {
	var contestID uint64
	ctx.BindID(&contestID)

	contest, err := s.ContestInteractor.Find(contestID)

	if err != nil {
		if err == usecases.ErrContestNotFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return fail.Wrap(err)
	}

	return ctx.JSON(http.StatusOK, contest)
}
