package domain

import (
	"time"
)

type mockClock struct {
	time time.Time
}

func (c *mockClock) Now() time.Time {
	return c.time
}

func (c *mockClock) SetTime(newTime time.Time) {
	c.time = newTime
}

func NewMockClock(time time.Time) *mockClock {
	return &mockClock{time: time}
}
