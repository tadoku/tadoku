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

//ErrNoRankingsFound for when you don't have any rankings to work with
var ErrNoRankingsFound = fail.New("no rankings found")

// ErrGlobalIsASystemLanguage for when you try to create a global ranking through the api
var ErrGlobalIsASystemLanguage = fail.New("global is a system language and cannot be created by the user")

// ErrInvalidContestLog for when an invalid contest is given
var ErrInvalidContestLog = fail.New("invalid contest log supplied")

// ErrContestLanguageNotSignedUp for when a user tries to log an entry for a contest with a language they're not signed up for
var ErrContestLanguageNotSignedUp = fail.New("user has not signed up for given language")

// RankingInteractor contains all business logic for rankings
type RankingInteractor interface {
	CreateRanking(
		contestID uint64,
		userID uint64,
		languages domain.LanguageCodes,
	) error
	CreateLog(log domain.ContestLog) error
	UpdateRanking(contestID uint64, userID uint64) error

	RankingsForContest(contestID uint64, languageCode domain.LanguageCode) (domain.Rankings, error)
}

// NewRankingInteractor instantiates RankingInteractor with all dependencies
func NewRankingInteractor(
	rankingRepository RankingRepository,
	contestRepository ContestRepository,
	contestLogRepository ContestLogRepository,
	userRepository UserRepository,
	validator Validator,
) RankingInteractor {
	return &rankingInteractor{
		rankingRepository:    rankingRepository,
		contestRepository:    contestRepository,
		contestLogRepository: contestLogRepository,
		userRepository:       userRepository,
		validator:            validator,
	}
}

type rankingInteractor struct {
	rankingRepository    RankingRepository
	contestRepository    ContestRepository
	contestLogRepository ContestLogRepository
	userRepository       UserRepository
	validator            Validator
}

func (i *rankingInteractor) CreateRanking(
	contestID uint64,
	userID uint64,
	languages domain.LanguageCodes,
) error {
	ids, err := i.contestRepository.GetOpenContests()
	if err != nil {
		return fail.Wrap(err)
	}

	if !domain.ContainsID(ids, contestID) {
		return ErrContestIsClosed
	}

	if _, err := i.userRepository.FindByID(userID); err != nil {
		return ErrUserDoesNotExist
	}

	existingLanguages, err := i.rankingRepository.GetAllLanguagesForContestAndUser(contestID, userID)
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
		if _, err := lang.Validate(); err != nil {
			return fail.Wrap(err)
		}

		ranking := domain.Ranking{
			ContestID: contestID,
			UserID:    userID,
			Language:  lang,
			Amount:    0,
		}
		err = i.rankingRepository.Store(ranking)
		if err != nil {
			return fail.Wrap(err)
		}
	}

	return nil
}

func (i *rankingInteractor) CreateLog(log domain.ContestLog) error {
	if valid, _ := i.validator.Validate(log); !valid {
		return ErrInvalidContestLog
	}

	ids, err := i.contestRepository.GetOpenContests()
	if err != nil {
		fail.Wrap(err)
	}
	if !domain.ContainsID(ids, log.ContestID) {
		return ErrContestIsClosed
	}

	languages, err := i.rankingRepository.GetAllLanguagesForContestAndUser(log.ContestID, log.UserID)
	if !languages.ContainsLanguage(log.Language) {
		return ErrContestLanguageNotSignedUp
	}

	err = i.contestLogRepository.Store(log)
	if err != nil {
		return fail.Wrap(err)
	}

	return i.UpdateRanking(log.ContestID, log.UserID)
}

func (i *rankingInteractor) UpdateRanking(contestID uint64, userID uint64) error {
	rankings, err := i.rankingRepository.FindAll(contestID, userID)
	if err != nil {
		return fail.Wrap(err)
	}

	if len(rankings) == 0 {
		return ErrNoRankingsFound
	}

	logs, err := i.contestLogRepository.FindAll(contestID, userID)
	if err != nil {
		return fail.Wrap(err)
	}

	totals := make(map[domain.LanguageCode]float32)
	for _, log := range logs {
		amount := log.AdjustedAmount()
		totals[log.Language] += amount
		totals[domain.Global] += amount
	}

	updatedRankings := domain.Rankings{}
	for _, ranking := range rankings {
		ranking.Amount = totals[ranking.Language]
		updatedRankings = append(updatedRankings, ranking)
	}

	return i.rankingRepository.UpdateAmounts(updatedRankings)
}

func (i *rankingInteractor) RankingsForContest(
	contestID uint64,
	languageCode domain.LanguageCode,
) (domain.Rankings, error) {

	validatedLanguage := domain.Global
	if ok, _ := i.validator.Validate(languageCode); ok {
		validatedLanguage = languageCode
	}

	lookup := i.rankingRepository.RankingsForContest
	if domain.ContestID(contestID).IsGlobal() {
		return nil, fail.New("not yet implemented")
	}

	rankings, err := lookup(contestID, validatedLanguage)
	if err != nil {
		return nil, fail.Wrap(err)
	}

	if len(rankings) == 0 {
		return nil, ErrNoRankingsFound
	}

	return rankings, nil
}
