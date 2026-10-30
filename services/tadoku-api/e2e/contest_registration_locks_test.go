package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestContestRegistrationEditWaitingPastContestEndIsRefused(t *testing.T) {
	const journey = "testdata/journeys/ContestRegistrationFrozenAfterEnd"
	lastSecond := time.Date(2026, 9, 30, 23, 59, 59, 0, time.UTC)
	dayAfterEnd := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	resetJourney(t, api, journey)
	tokens, err := loadCast([]step{{as: user}})
	if err != nil {
		t.Fatal(err)
	}
	previous := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time { return fixtureInstant }
	defer func() { jwt.TimeFunc = previous }()
	request := readHTTPRequest(t, journey+"/01_remove_language_after_end")
	authorize(request, user, tokens)
	holder := holdRow(t, "select 1 from users where id = $1 for update", readerUserID)

	// Test safety: the request starts on the contest's last second and stays parked on the held users row while the
	// clock moves past midnight, so it reads business time only inside the two scopes.
	var served <-chan *httptest.ResponseRecorder
	timex.TheWorld(lastSecond, func() {
		served = serveAsync(request)
		waitForLockWaiters(t, 1, served)
	})
	timex.TheWorld(dayAfterEnd, func() {
		if err := holder.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireBody(t, <-served, http.StatusForbidden, "")
	})

	checkVerifyGolden(t, api, journey+"/03_logs_and_registration_kept")
}
