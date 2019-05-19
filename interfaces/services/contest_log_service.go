package services

import (
	"net/http"
	"strconv"

	"github.com/srvc/fail"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// ContestLogService is responsible for managing contest log entries
type ContestLogService interface {
	Create(ctx Context) error
	Get(ctx Context) error
}

// NewContestLogService initializer
func NewContestLogService(rankingInteractor usecases.RankingInteractor) ContestLogService {
	return &contestLogService{
		RankingInteractor: rankingInteractor,
	}
}

type contestLogService struct {
	RankingInteractor usecases.RankingInteractor
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

	if err := s.RankingInteractor.CreateLog(*log); err != nil {
		return fail.Wrap(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

func (s *contestLogService) Get(ctx Context) error {
	contestID, err := strconv.ParseUint(ctx.QueryParam("contest_id"), 10, 64)
	if err != nil {
		return fail.Wrap(err)
	}

	userID, err := strconv.ParseUint(ctx.QueryParam("user_id"), 10, 64)
	if err != nil {
		return fail.Wrap(err)
	}

	logs, err := s.RankingInteractor.ContestLogs(contestID, userID)
	if err != nil {
		if err == usecases.ErrNoContestLogsFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return fail.Wrap(err)
	}

	return ctx.JSON(http.StatusOK, logs)
}
