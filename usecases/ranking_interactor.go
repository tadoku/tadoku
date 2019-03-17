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

// ErrContestIsClosed for when you try to log for a closed contest
var ErrContestIsClosed = fail.New("the given contest is closed")

// ErrNoRankingToCreate for when you try to create a ranking that already exists
var ErrNoRankingToCreate = fail.New("there is no new ranking to be created")

// ErrGlobalIsASystemLanguage for when you try to create a global ranking through the api
var ErrGlobalIsASystemLanguage = fail.New("global is a system language and cannot be created by the user")

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
	contestRepository ContestRepository,
	userRepository UserRepository,
	validator Validator,
) RankingInteractor {
	return &rankingInteractor{
		rankingRepository: rankingRepository,
		contestRepository: contestRepository,
		userRepository:    userRepository,
		validator:         validator,
	}
}

type rankingInteractor struct {
	rankingRepository RankingRepository
	contestRepository ContestRepository
	userRepository    UserRepository
	validator         Validator
}

func (si *rankingInteractor) CreateRanking(
	userID uint64,
	contestID uint64,
	languages domain.LanguageCodes,
) error {
	ids, err := si.contestRepository.GetOpenContests()
	if err != nil {
		return fail.Wrap(err)
	}

	if !domain.ContainsID(ids, contestID) {
		return ErrContestIsClosed
	}

	if _, err := si.userRepository.FindByID(userID); err != nil {
		return ErrUserDoesNotExist
	}

	existingLanguages, err := si.rankingRepository.GetAllLanguagesForContestAndUser(contestID, userID)
	if err != nil {
		return fail.Wrap(err)
	}
	needsGlobal := !existingLanguages.ContainsLanguage(domain.Global)

	// Figure out which languages we need to create new rankings for
	targetLanguages := domain.LanguageCodes{}
	for _, lang := range languages {
		if lang == domain.Global {
			return ErrGlobalIsASystemLanguage
		}

		if existingLanguages.ContainsLanguage(lang) {
			continue
		}
		targetLanguages = append(targetLanguages, lang)
	}

	if needsGlobal {
		targetLanguages = append(targetLanguages, domain.Global)
	}

	if len(targetLanguages) == 0 {
		return ErrNoRankingToCreate
	}

	for _, lang := range targetLanguages {
		ranking := domain.Ranking{
			ContestID: contestID,
			UserID:    userID,
			Language:  lang,
			Amount:    0,
		}
		err = si.rankingRepository.Store(ranking)
		if err != nil {
			return fail.Wrap(err)
		}
	}

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
