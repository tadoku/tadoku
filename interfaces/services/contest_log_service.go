package services

import (
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
	ctx.NoContent(201)

	return nil
}
