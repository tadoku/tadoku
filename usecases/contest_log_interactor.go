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
func NewContestLogInteractor(contestLogRepository ContestLogRepository) ContestLogInteractor {
	return &contestLogInteractor{
		contestLogRepository: contestLogRepository,
	}
}

type contestLogInteractor struct {
	contestLogRepository ContestLogRepository
}

func (si *contestLogInteractor) CreateLog(log domain.ContestLog) error {
	return fail.Errorf("not implemented yet")
}
