//go:generate gex mockgen -source=contest_log_interactor.go -package usecases -destination=contest_log_interactor_mock.go

package usecases

import (
	"github.com/srvc/fail"

	"github.com/tadoku/api/domain"
)

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
) ContestLogInteractor {
	return &contestLogInteractor{
		contestLogRepository: contestLogRepository,
		contestRepository:    contestRepository,
		rankingRepository:    rankingRepository,
	}
}

type contestLogInteractor struct {
	contestLogRepository ContestLogRepository
	contestRepository    ContestRepository
	rankingRepository    RankingRepository
}

func (i *contestLogInteractor) CreateLog(log domain.ContestLog) error {
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
	// TODO: create log
	// TODO: recalculate rankings

	return i.contestLogRepository.Store(log)
}
