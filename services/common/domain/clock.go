package domain

import (
	"time"

	"github.com/pkg/errors"
)

// Clock abstracts away getting the current time, so it can be mocked in tests
type Clock interface {
	Now() time.Time
}

type realTimeClock struct {
	location *time.Location
}

// Now gets the current time in UTC
func (c *realTimeClock) Now() time.Time {
	now := time.Now().In(c.location)

	return now
}

// New creates a new clock for a given location
func NewClock(locationName string) (Clock, error) {
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		return nil, errors.Wrap(err, "could not create location for real time clock")
	}

	return &realTimeClock{location: loc}, nil
}
