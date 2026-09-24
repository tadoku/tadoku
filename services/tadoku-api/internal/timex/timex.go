package timex

import (
	"sync/atomic"
	"time"
)

var fixedTime atomic.Pointer[time.Time]

func Now() time.Time {
	if fixed := fixedTime.Load(); fixed != nil {
		return *fixed
	}
	return time.Now().UTC().Truncate(time.Microsecond)
}

// TheWorld changes time process-wide. Tests using it and their parents must
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
