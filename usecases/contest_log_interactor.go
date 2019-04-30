//go:generate gex mockgen -source=contest_log_interactor.go -package usecases -destination=contest_log_interactor_mock.go

package usecases

import (
	"github.com/tadoku/api/domain"
)

// ContestLogInteractor contains all business logic for contest log entries
type ContestLogInteractor interface {
	CreateLog(log domain.Log) error
}

// NewContestLogInteractor instantiates ContestLogInteractor with all dependencies
func NewContestInteractor() ContestLogInteractor {
	return &contestLogInteractor{}
}

type contestLogInteractor struct {
}

func (si *contestInteractor) CreateLog(log domain.Log) error {
	return nil
}
