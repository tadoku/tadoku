//go:generate gex mockgen -source=contest_interactor.go -package usecases -destination=contest_interactor_mock.go

package usecases

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
)

// ContestInteractor contains all business logic for contests
type ContestInteractor interface {
	CreateContest(contest domain.Contest) error
}

// NewContestInteractor instantiates ContestInteractor with all dependencies
func NewContestInteractor(
	contestRepository ContestRepository,
) ContestInteractor {
	return &contestInteractor{
		contestRepository: contestRepository,
	}
}

type contestInteractor struct {
	contestRepository ContestRepository
}

func (si *contestInteractor) CreateContest(contest domain.Contest) error {
	if contest.ID != 0 {
		return fail.Errorf("user with an id (%v) could not be created", contest.ID)
	}

	// Add validation here
	// if err := si.DomainValidator.Validate(contest); err != nil {
	//   return fail.Errorf("contest is invalid")
	// }

	err := si.contestRepository.Store(contest)
	return fail.Wrap(err)
}
