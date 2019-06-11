package domain

import (
	"time"

	"github.com/srvc/fail"
)

// Contest contains the data about when a contest is being held
type Contest struct {
	ID          uint64    `json:"id" db:"id"`
	Description string    `json:"description" db:"description" valid:"required"`
	Start       time.Time `json:"start" db:"start" valid:"required"`
	End         time.Time `json:"end" db:"end" valid:"required"`
	Open        bool      `json:"open" db:"open"`
}

// Contests is a collection of contests
type Contests []Contest

// ErrContestInvalidDateOrder for when you accidentally switch up the starting and ending dates
var ErrContestInvalidDateOrder = fail.New("contest must start before it can end")

// ErrContestInvalidDateTooOld for when you try to make a contest that has already ended
var ErrContestInvalidDateTooOld = fail.New("contest must end in the future")

// Validate a contest
func (c Contest) Validate() (bool, error) {
	if c.Start.After(c.End) {
		return false, ErrContestInvalidDateOrder
	}
	if c.End.Before(time.Now()) {
		return false, ErrContestInvalidDateTooOld
	}

	return true, nil
}

// ContestID is a container for contest ids with some domain logic
type ContestID uint64

// IsGlobal for knowing if a contest is a specific one or all of them
func (id ContestID) IsGlobal() bool {
	return id == 0
}
