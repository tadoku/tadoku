package leaderboard

import (
	"strconv"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestYearlyRequestValidateRejectsOutOfRangeYears(t *testing.T) {
	t.Parallel()
	const want = "year must be between 1 and 9999"

	for _, year := range []int{0, 10000, 1<<32 + 2026} {
		t.Run(strconv.Itoa(year), func(t *testing.T) {
			err := YearlyRequest{Year: year}.Validate()
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != want {
				t.Errorf("Validate() error = %v, want invalid input %q", err, want)
			}
		})
	}

	t.Run("2026", func(t *testing.T) {
		if err := (YearlyRequest{Year: 2026}).Validate(); err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})
}
