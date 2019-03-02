//go:generate gex mockgen -source=contest_interactor.go -package usecases -destination=contest_interactor_mock.go

package usecases

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
)

// ErrInvalidContest for when an invalid contest is given
var ErrInvalidContest = fail.New("invalid contest supplied")

// ContestInteractor contains all business logic for contests
type ContestInteractor interface {
	CreateContest(contest domain.Contest) error
}

// NewContestInteractor instantiates ContestInteractor with all dependencies
func NewContestInteractor(
	contestRepository ContestRepository,
	validator Validator,
) ContestInteractor {
	return &contestInteractor{
		contestRepository: contestRepository,
		validator:         validator,
	}
}

type contestInteractor struct {
	contestRepository ContestRepository
	validator         Validator
}

func (si *contestInteractor) CreateContest(contest domain.Contest) error {
	if contest.ID != 0 {
		return fail.Errorf("user with an id (%v) could not be created", contest.ID)
	}

	if _, err := si.validator.Validate(contest); err != nil {
		return ErrInvalidContest
	}

	err := si.contestRepository.Store(contest)
	return fail.Wrap(err)
}
