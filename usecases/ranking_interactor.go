//go:generate gex mockgen -source=ranking_interactor.go -package usecases -destination=ranking_interactor_mock.go

package usecases

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
)

// ErrInvalidRanking for when an invalid ranking is given
var ErrInvalidRanking = fail.New("invalid ranking supplied")

// ErrRankingIDMissing for when you try to update a ranking without id
var ErrRankingIDMissing = fail.New("a ranking id is required when updating")

// RankingInteractor contains all business logic for rankings
type RankingInteractor interface {
	CreateRanking(
		userID uint64,
		contestID uint64,
		languages domain.LanguageCodes,
	) error
	UpdateRanking(ranking domain.Ranking) error
}

// NewRankingInteractor instantiates RankingInteractor with all dependencies
func NewRankingInteractor(
	rankingRepository RankingRepository,
	validator Validator,
) RankingInteractor {
	return &rankingInteractor{
		rankingRepository: rankingRepository,
		validator:         validator,
	}
}

type rankingInteractor struct {
	rankingRepository RankingRepository
	validator         Validator
}

func (si *rankingInteractor) CreateRanking(
	userID uint64,
	contestID uint64,
	languages domain.LanguageCodes,
) error {
	// check if contest is open

	// check if user exists

	// check if user has rankings for a certain language already

	return nil
}

func (si *rankingInteractor) UpdateRanking(ranking domain.Ranking) error {
	if ranking.ID == 0 {
		return ErrRankingIDMissing
	}

	if _, err := si.validator.ValidateStruct(ranking); err != nil {
		return ErrInvalidRanking
	}

	// TODO: Check if contest is open

	err := si.rankingRepository.Store(ranking)
	return fail.Wrap(err)
}
