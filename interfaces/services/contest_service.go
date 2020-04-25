package services

import (
	"net/http"
	"strconv"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// ContestService is responsible for managing contests
type ContestService interface {
	Create(ctx Context) error
	Update(ctx Context) error
	All(ctx Context) error
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
		return domain.WrapError(err)
	}

	if err := s.ContestInteractor.CreateContest(*contest); err != nil {
		return domain.WrapError(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

func (s *contestService) Update(ctx Context) error {
	contest := &domain.Contest{}
	if err := ctx.Bind(contest); err != nil {
		return domain.WrapError(err)
	}

	ctx.BindID(&contest.ID)

	if err := s.ContestInteractor.UpdateContest(*contest); err != nil {
		return domain.WrapError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (s *contestService) All(ctx Context) error {
	limit, err := strconv.ParseInt(ctx.QueryParam("limit"), 10, 64)
	if err != nil {
		limit = 0
	}
	contests, err := s.ContestInteractor.Recent(int(limit))

	if err != nil {
		if err == usecases.ErrContestNotFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return domain.WrapError(err)
	}

	return ctx.JSON(http.StatusOK, contests)
}

func (s *contestService) Get(ctx Context) error {
	var contestID uint64
	ctx.BindID(&contestID)

	contest, err := s.ContestInteractor.Find(contestID)

	if err != nil {
		if err == usecases.ErrContestNotFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return domain.WrapError(err)
	}

	return ctx.JSON(http.StatusOK, contest)
}
