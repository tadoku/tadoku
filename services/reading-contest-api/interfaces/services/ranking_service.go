package services

import (
	"net/http"
	"strconv"

	"github.com/tadoku/tadoku/services/reading-contest-api/domain"
	"github.com/tadoku/tadoku/services/reading-contest-api/usecases"
)

// RankingService is responsible for managing rankings
type RankingService interface {
	Get(ctx Context) error
	CurrentRegistration(ctx Context) error
	RankingsForRegistration(ctx Context) error
}

// NewRankingService initializer
func NewRankingService(rankingInteractor usecases.RankingInteractor) RankingService {
	return &rankingService{
		RankingInteractor: rankingInteractor,
	}
}

type rankingService struct {
	RankingInteractor usecases.RankingInteractor
}

func (s *rankingService) Get(ctx Context) error {
	contestID, err := strconv.ParseUint(ctx.QueryParam("contest_id"), 10, 64)
	if err != nil {
		return domain.WrapError(err)
	}

	rankings, err := s.RankingInteractor.RankingsForContest(contestID)
	if err != nil {
		if err == usecases.ErrNoRankingsFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return domain.WrapError(err)
	}

	return ctx.JSON(http.StatusOK, rankings.GetView())
}

func (s *rankingService) CurrentRegistration(ctx Context) error {
	user, err := ctx.User()
	if err != nil {
		return domain.WrapError(err)
	}

	registration, err := s.RankingInteractor.CurrentRegistration(user.ID)
	if err != nil {
		if err == usecases.ErrNoRankingRegistrationFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return domain.WrapError(err)
	}

	return ctx.JSON(http.StatusOK, registration)
}

func (s *rankingService) RankingsForRegistration(ctx Context) error {
	contestID, err := strconv.ParseUint(ctx.QueryParam("contest_id"), 10, 64)
	if err != nil {
		return domain.WrapError(err)
	}
	userID, err := strconv.ParseUint(ctx.QueryParam("user_id"), 10, 64)
	if err != nil {
		return domain.WrapError(err)
	}

	rankings, err := s.RankingInteractor.RankingsForRegistration(contestID, userID)
	if err != nil {
		if err == usecases.ErrNoRankingsFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return domain.WrapError(err)
	}

	return ctx.JSON(http.StatusOK, rankings.GetView())
}
