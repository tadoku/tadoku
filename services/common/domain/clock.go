package domain

import (
	"fmt"
	"time"
)

type Clock interface {
	Now() time.Time
}

type realTimeClock struct {
	location *time.Location
}

func (c *realTimeClock) Now() time.Time {
	return time.Now().In(c.location)
}

func NewClock(locationName string) (Clock, error) {
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		return nil, fmt.Errorf("could not create location for real time clock: %w", err)
	}

	return &realTimeClock{location: loc}, nil
}
