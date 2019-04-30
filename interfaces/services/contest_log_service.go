package services

import (
	"github.com/srvc/fail"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// ContestLogService is responsible for managing contest log entries
type ContestLogService interface {
	Create(ctx Context) error
}

// NewContestLogService initializer
func NewContestLogService(contestLogInteractor usecases.ContestLogInteractor) ContestLogService {
	return &contestLogService{
		ContestLogInteractor: contestLogInteractor,
	}
}

type contestLogService struct {
	ContestLogInteractor usecases.ContestLogInteractor
}

func (s *contestLogService) Create(ctx Context) error {
	log := &domain.ContestLog{}
	if err := ctx.Bind(log); err != nil {
		return fail.Wrap(err)
	}

	user, err := ctx.User()
	if err != nil {
		return fail.Wrap(err)
	}

	log.UserID = user.ID

	if err := s.ContestLogInteractor.CreateLog(*log); err != nil {
		return fail.Wrap(err)
	}

	return ctx.NoContent(201)
}
