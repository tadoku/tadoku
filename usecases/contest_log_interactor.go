//go:generate gex mockgen -source=contest_log_interactor.go -package usecases -destination=contest_log_interactor_mock.go

package usecases

import (
	"github.com/srvc/fail"

	"github.com/tadoku/api/domain"
)

// ErrInvalidContestLog for when an invalid contest is given
var ErrInvalidContestLog = fail.New("invalid contest log supplied")

// ErrContestLanguageNotSignedUp for when a user tries to log an entry for a contest with a langauge they're not signed up for
var ErrContestLanguageNotSignedUp = fail.New("user has not signed up for given language")

// ContestLogInteractor contains all business logic for contest log entries
type ContestLogInteractor interface {
	CreateLog(log domain.ContestLog) error
}

// NewContestLogInteractor instantiates ContestLogInteractor with all dependencies
func NewContestLogInteractor(
	contestLogRepository ContestLogRepository,
	contestRepository ContestRepository,
	rankingRepository RankingRepository,
	validator Validator,
) ContestLogInteractor {
	return &contestLogInteractor{
		contestLogRepository: contestLogRepository,
		contestRepository:    contestRepository,
		rankingRepository:    rankingRepository,
		validator:            validator,
	}
}

type contestLogInteractor struct {
	contestLogRepository ContestLogRepository
	contestRepository    ContestRepository
	rankingRepository    RankingRepository
	validator            Validator
}

func (i *contestLogInteractor) CreateLog(log domain.ContestLog) error {
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

	// TODO: recalculate rankings
	return nil
}
