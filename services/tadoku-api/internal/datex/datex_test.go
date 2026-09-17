package datex_test

import (
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/datex"
)

func TestIsValidRange(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name  string
		start time.Time
		end   time.Time
		want  bool
	}{
		{name: "ordered", start: start, end: start.Add(time.Hour), want: true},
		{name: "nanosecond", start: start, end: start.Add(time.Nanosecond), want: true},
		{name: "missing start", end: start},
		{name: "missing end", start: start},
		{name: "both missing"},
		{name: "equal", start: start, end: start},
		{name: "reversed", start: start, end: start.Add(-time.Hour)},
		{name: "same instant different zone", start: start, end: start.In(time.FixedZone("offset", 2*60*60))},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := datex.IsValidRange(test.start, test.end); got != test.want {
				t.Errorf("IsValidRange(%v, %v) = %t, want %t", test.start, test.end, got, test.want)
			}
		})
	}
}
