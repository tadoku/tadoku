package logs

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestYearlyStatisticsRejectOutOfRangeYears(t *testing.T) {
	t.Parallel()
	service := NewService(nil, false)
	userID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	const want = "year must be between 1 and 9999"

	for _, year := range []int{0, 10000, 40000} {
		t.Run(strconv.Itoa(year), func(t *testing.T) {
			_, err := service.YearlyActivity(t.Context(), userID, year)
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != want {
				t.Errorf("YearlyActivity: error=%v, want invalid input %q", err, want)
			}

			_, err = service.YearlyScores(t.Context(), userID, year)
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != want {
				t.Errorf("YearlyScores: error=%v, want invalid input %q", err, want)
			}

			_, err = service.YearlyActivitySplit(t.Context(), userID, year)
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != want {
				t.Errorf("YearlyActivitySplit: error=%v, want invalid input %q", err, want)
			}
		})
	}
}
