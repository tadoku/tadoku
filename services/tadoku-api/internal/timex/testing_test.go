// Package timex_test provides test-only control over the Tadoku API business clock.
package timex_test

import (
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex/internal/clockstate"
)

// TheWorld calls fn while Now returns instant in UTC for all goroutines in the
// process. Normal time is restored when fn returns or panics; a panic propagates
// unchanged. A nested or concurrent override panics without calling its callback.
//
// Tests using TheWorld, and their parent tests, must not use t.Parallel. All
// participating goroutines must finish before fn returns. Use real time for
// timeouts, context deadlines, tickers, sleeping, and elapsed measurements.
func TheWorld(instant time.Time, fn func()) {
	if !clockstate.Set(instant.UTC()) {
		panic("timex_test.TheWorld: override already active")
	}
	defer clockstate.Clear()
	fn()
}
