package domain

import "time"

// Contest contains the data about when a contest is being held
type Contest struct {
	ID    uint64    `json:"id" db:"id"`
	Start time.Time `json:"start" db:"start" valid:"required"`
	End   time.Time `json:"end" db:"end" valid:"required"`
	Open  bool      `json:"open" db:"open" valid:"required"`
}

// Contests is a collection of contests
type Contests []Contest
