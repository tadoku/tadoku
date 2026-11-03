package timex

import (
	"context"
	"sync"
	"time"
)

// Clock is a frozen business instant that a test sets and advances. Attach it with WithClock.
type Clock struct {
	mu      sync.Mutex
	instant time.Time
}

func NewClock(instant time.Time) *Clock {
	return &Clock{instant: instant.UTC()}
}

func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.instant
}

func (c *Clock) Set(instant time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.instant = instant.UTC()
}

func (c *Clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.instant = c.instant.Add(d)
}

type clockKey struct{}

// WithClock makes Now read clock for ctx and every context derived from it. Production code never attaches one.
func WithClock(ctx context.Context, clock *Clock) context.Context {
	return context.WithValue(ctx, clockKey{}, clock)
}

func contextClock(ctx context.Context) (*Clock, bool) {
	clock, ok := ctx.Value(clockKey{}).(*Clock)
	return clock, ok && clock != nil
}
