package domain

import (
	"time"
)

type mockClock struct {
	time time.Time
}

// Now gets the current time of the mocked clock
func (c *mockClock) Now() time.Time {
	return c.time
}

func (c *mockClock) SetTime(newTime time.Time) {
	c.time = newTime
}

// NewMock creates a clock that always returns the same time
func NewMockClock(time time.Time) *mockClock {
	return &mockClock{time: time}
}
