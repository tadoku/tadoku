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

// ErrContestIDMissing for when you try to update a contest without id
var ErrContestIDMissing = fail.New("a contest id is required when updating")

// ErrCreateContestHasID for when you try to create a contest with a given id
var ErrCreateContestHasID = fail.New("a contest can't have an id when being created")

// ContestInteractor contains all business logic for contests
type ContestInteractor interface {
	CreateContest(contest domain.Contest) error
	UpdateContest(contest domain.Contest) error
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
		return ErrCreateContestHasID
	}

	return si.saveContest(contest)
}

func (si *contestInteractor) UpdateContest(contest domain.Contest) error {
	if contest.ID == 0 {
		return ErrContestIDMissing
	}

	return si.saveContest(contest)
}

func (si *contestInteractor) saveContest(contest domain.Contest) error {
	if valid, _ := si.validator.Validate(contest); !valid {
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

	err := si.contestRepository.Store(&contest)
	return fail.Wrap(err)
}
