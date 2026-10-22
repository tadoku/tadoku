package logs

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
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

func TestRegistrationsForRescoringIncludesContestEndDay(t *testing.T) {
	t.Parallel()

	registrationID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	contestID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	contestEnd := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	log := &Log{
		Registrations: []RegistrationReference{{
			RegistrationID: registrationID,
			ContestID:      contestID,
			ContestEnd:     contestEnd,
		}},
	}
	want := []logscore.Target{{
		RegistrationID: registrationID,
		ContestID:      contestID,
	}}

	duringEndDay := NewService(nil, true).RegistrationsForRescoring(
		log,
		time.Date(2026, 9, 12, 15, 0, 0, 0, time.UTC),
	)
	if len(duringEndDay) != 1 || duringEndDay[0] != want[0] {
		t.Errorf("during end day: got %#v, want %#v", duringEndDay, want)
	}

	afterEndDay := NewService(nil, true).RegistrationsForRescoring(
		log,
		time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
	)
	if len(afterEndDay) != 0 {
		t.Errorf("after end day: got %#v, want none", afterEndDay)
	}

	disabled := NewService(nil, false).RegistrationsForRescoring(
		log,
		time.Date(2026, 9, 12, 15, 0, 0, 0, time.UTC),
	)
	if disabled != nil {
		t.Errorf("scoring engine disabled: got %#v, want nil", disabled)
	}
}
