//go:generate gex mockgen -source=contest_interactor.go -package usecases -destination=contest_interactor_mock.go

package usecases

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
)

// ErrInvalidContest for when an invalid contest is given
var ErrInvalidContest = fail.New("invalid contest supplied")

// ErrOpenContestAlreadyExists for when you try to create a second contest that's open
var ErrOpenContestAlreadyExists = fail.New("an open contest already exists, only one can exist at a time")

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

	if contest.Open {
		ids, err := si.contestRepository.GetOpenContests()
		if err != nil {
			return fail.Wrap(err)
		}
		if len(ids) > 0 {
			for _, id := range ids {
				if id != contest.ID {
					return ErrOpenContestAlreadyExists
				}
			}
		}
	}

	err := si.contestRepository.Store(contest)
	return fail.Wrap(err)
}
