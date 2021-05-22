package infra

import (
	"time"

	"github.com/srvc/fail"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// NewClock creates a new clock for a given location
func NewClock(locationName string) (usecases.Clock, error) {
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		return nil, domain.WrapError(err, fail.WithMessage("could not create location for real time clock"))
	}

	return &realTimeClock{location: loc}, nil
}

type realTimeClock struct {
	location *time.Location
}

// Now gets the current time in UTC
func (c *realTimeClock) Now() time.Time {
	now := time.Now().In(c.location)

	return now
}
