package leaderboard

import (
	"strconv"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestYearlyRequestValidateRejectsOutOfRangeYears(t *testing.T) {
	t.Parallel()
	for _, year := range []int32{0, 10000, 40000} {
		t.Run(strconv.Itoa(int(year)), func(t *testing.T) {
			err := YearlyRequest{Year: year}.Validate()
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != "year must be between 1 and 9999" {
				t.Errorf("error=%v, want invalid input %q", err, "year must be between 1 and 9999")
			}
		})
	}
}
