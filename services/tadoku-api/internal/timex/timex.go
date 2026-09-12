// Package timex provides UTC wall-clock time and a scoped, process-global
// override for tests. Overrides do not affect timers, deadlines, or database time.
package timex

import (
	"sync/atomic"
	"time"
)

var fixedTime atomic.Pointer[time.Time]

// Now returns the current wall-clock time in UTC, or the instant supplied to an
// active TheWorld callback.
func Now() time.Time {
	if fixed := fixedTime.Load(); fixed != nil {
		return *fixed
	}
	return time.Now().UTC()
}

// TheWorld calls fn while Now returns instant in UTC for all goroutines in the
// process. Normal time is restored when fn returns or panics; a panic propagates
// unchanged. A nested or concurrent override panics with
// "timex.TheWorld: override already active" without calling its callback.
//
// Tests using TheWorld, and their parent tests, must not use t.Parallel. All
// participating goroutines must finish before fn returns. Use real time for
// timeouts, context deadlines, tickers, sleeping, and elapsed measurements.
func TheWorld(instant time.Time, fn func()) {
	fixed := instant.UTC()
	if !fixedTime.CompareAndSwap(nil, &fixed) {
		panic("timex.TheWorld: override already active")
	}
	defer fixedTime.Store(nil)
	fn()
}
