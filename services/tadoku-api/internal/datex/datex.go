// Package datex provides shared date and time validation.
package datex

import "time"

// IsValidRange reports whether both endpoints are set and start precedes end.
func IsValidRange(start, end time.Time) bool {
	return !start.IsZero() && !end.IsZero() && start.Before(end)
}
