//go:generate gex mockgen -source=contest_log_interactor.go -package usecases -destination=contest_log_interactor_mock.go

package usecases

import (
	"github.com/srvc/fail"

	"github.com/tadoku/api/domain"
)

// ContestLogInteractor contains all business logic for contest log entries
type ContestLogInteractor interface {
	CreateLog(log domain.ContestLog) error
}

// NewContestLogInteractor instantiates ContestLogInteractor with all dependencies
func NewContestLogInteractor(
	contestLogRepository ContestLogRepository,
	contestRepository ContestRepository,
) ContestLogInteractor {
	return &contestLogInteractor{
		contestLogRepository: contestLogRepository,
		contestRepository:    contestRepository,
	}
}

type contestLogInteractor struct {
	contestLogRepository ContestLogRepository
	contestRepository    ContestRepository
}

func (i *contestLogInteractor) CreateLog(log domain.ContestLog) error {
	ids, err := i.contestRepository.GetOpenContests()
	if err != nil {
		fail.Wrap(err)
	}
	if !domain.ContainsID(ids, log.ContestID) {
		return ErrContestIsClosed
	}

	// TODO: Check if signed up for given language
	// TODO: create log
	// TODO: recalculate rankings

	return i.contestLogRepository.Store(log)
}
