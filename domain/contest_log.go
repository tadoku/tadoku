package domain

import (
	"time"
)

// ContestLog contains a single entry in a contest for a user
type ContestLog struct {
	ID        uint64       `json:"id" db:"id"`
	ContestID uint64       `json:"contest_id" db:"contest_id" valid:"required"`
	UserID    uint64       `json:"user_id" db:"user_id" valid:"required"`
	Language  LanguageCode `json:"language_code" db:"language_code" valid:"required"`
	MediumID  MediumID     `json:"medium_id" db:"medium_id" valid:"required"`
	Amount    float32      `json:"amount" db:"amount" valid:"required"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at" db:"deleted_at"`
}

// ContestLogs is a collection of ContestLog
type ContestLogs []ContestLog

// Validate a contest log
func (c ContestLog) Validate() (bool, error) {
	if valid, err := c.MediumID.Validate(); !valid {
		return valid, err
	}

	return true, nil
}

// AdjustedAmount gives the amount after having taken into account the medium
func (c ContestLog) AdjustedAmount() float32 {
	return c.MediumID.AdjustedAmount(c.Amount)
}
