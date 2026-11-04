package leaderboard

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestInvalidateContestRequiresTenant(t *testing.T) {
	t.Parallel()

	err := NewService(nil, NewCache(nil, 0)).InvalidateContest(t.Context(), uuid.New())
	if errx.KindOf(err) != errx.Unavailable {
		t.Errorf("InvalidateContest without tenant = %v, want unavailable", err)
	}
}

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

func TestInvalidateOfficialRejectsOutOfRangeYears(t *testing.T) {
	t.Parallel()
	service := NewService(nil, nil)
	for _, year := range []int16{0, 10000} {
		t.Run(strconv.Itoa(int(year)), func(t *testing.T) {
			err := service.InvalidateOfficial(t.Context(), year)
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != "year must be between 1 and 9999" {
				t.Errorf("error=%v, want invalid input %q", err, "year must be between 1 and 9999")
			}
		})
	}
}
