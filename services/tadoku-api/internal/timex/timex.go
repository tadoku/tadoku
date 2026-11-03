package timex

import (
	"context"
	"sync/atomic"
	"time"
)

var fixedTime atomic.Pointer[time.Time]

// Now returns business time: the clock attached to ctx with WithClock, else an active TheWorld instant, else the real
// UTC time at microsecond precision.
func Now(ctx context.Context) time.Time {
	if clock, ok := contextClock(ctx); ok {
		return clock.Now()
	}
	if fixed := fixedTime.Load(); fixed != nil {
		return *fixed
	}
	return time.Now().UTC().Truncate(time.Microsecond)
}

// TheWorld changes time process-wide for contexts without a clock. Tests using it and their parents must
// not use t.Parallel; all participating goroutines must finish before fn returns.
// Use real time for timeouts, deadlines, tickers, sleeping and elapsed measurements.
func TheWorld(instant time.Time, fn func()) {
	fixed := instant.UTC()
	if !fixedTime.CompareAndSwap(nil, &fixed) {
		panic("timex.TheWorld: override already active")
	}
	defer fixedTime.Store(nil)
	fn()
}

// Advance moves the time an active TheWorld scope fixes, so a test can let business time pass while a request is in
// flight. It panics outside a scope.
func Advance(d time.Duration) {
	for {
		fixed := fixedTime.Load()
		if fixed == nil {
			panic("timex.Advance: no override active")
		}

		next := fixed.Add(d)
		if fixedTime.CompareAndSwap(fixed, &next) {
			return
		}
	}
}
