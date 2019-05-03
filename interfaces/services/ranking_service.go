package services

import (
	"net/http"

	"github.com/srvc/fail"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// RankingService is responsible for managing rankings
type RankingService interface {
	Create(ctx Context) error
	Get(ctx Context) error
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

// CreateRankingPayload payload for the create action
type CreateRankingPayload struct {
	ContestID uint64               `json:"contest_id"`
	Languages domain.LanguageCodes `json:"languages"`
}

func (s *rankingService) Create(ctx Context) error {
	payload := &CreateRankingPayload{}
	if err := ctx.Bind(payload); err != nil {
		return fail.Wrap(err)
	}

	user, err := ctx.User()
	if err != nil {
		return fail.Wrap(err)
	}

	if err := s.RankingInteractor.CreateRanking(payload.ContestID, user.ID, payload.Languages); err != nil {
		return fail.Wrap(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

func (s *rankingService) Get(ctx Context) error {
	contestID, err := ctx.GetID()
	if err != nil {
		return fail.Wrap(err)
	}

	rankings, err := s.RankingInteractor.RankingsForContest(contestID, domain.Global)
	if err != nil {
		if err == usecases.ErrNoRankingsFound {
			return ctx.NoContent(http.StatusNotFound)
		}

		return fail.Wrap(err)
	}

	return ctx.JSON(http.StatusOK, rankings)
}
