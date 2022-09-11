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

type ContestStats struct {
	ByLanguage   []ContestLanguageStat `json:"by_language"`
	Participants int                   `json:"participants"`
	TotalAmount  float64               `json:"total_amount"`
}

type ContestLanguageStat struct {
	Count        int    `json:"count" db:"cnt"`
	LanguageCode string `json:"language_code" db:"language_code"`
}

// ContestRegistration holds the contest registration data for a user
type ContestRegistration struct {
	ID              uint64        `json:"id" db:"id"`
	UserID          uint64        `json:"user_id" db:"user_id" valid:"required"`
	UserDisplayName string        `json:"user_display_name" db:"user_display_name"`
	ContestID       uint64        `json:"contest_id" db:"contest_id" valid:"required"`
	LanguageCodes   LanguageCodes `json:"languages" db:"language_codes" valid:"required"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}
