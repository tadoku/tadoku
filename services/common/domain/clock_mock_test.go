package domain_test

import (
	"time"

	"github.com/tadoku/tadoku/services/common/domain"
)

type mockClock struct {
	time time.Time
}

// Now gets the current time of the mocked clock
func (c *mockClock) Now() time.Time {
	return c.time
}

// NewMock creates a clock that always returns the same time
func NewClockMock(time time.Time) domain.Clock {
	return &mockClock{time: time}
}
