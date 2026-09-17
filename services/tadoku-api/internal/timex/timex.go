// Package timex provides UTC wall-clock time.
package timex

import (
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex/internal/clockstate"
)

// Now returns the current wall-clock time in UTC.
func Now() time.Time {
	if fixed, ok := clockstate.Fixed(); ok {
		return fixed
	}
	return time.Now().UTC().Truncate(time.Microsecond)
}
