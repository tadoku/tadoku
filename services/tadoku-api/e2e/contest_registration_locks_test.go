package e2e_test

import (
	"context"
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

	timex.TheWorld(lastSecond, func() {
		served := make(chan *httptest.ResponseRecorder, 1)
		finished := make(chan struct{})
		go func() {
			defer close(finished)
			response := httptest.NewRecorder()
			api.handler.ServeHTTP(response, request)
			served <- response
		}()
		// Test safety: releasing the row and waiting for the request here keeps it inside this scope on every path,
		// including a failed lock wait.
		defer func() {
			_ = holder.Rollback(context.Background())
			<-finished
		}()

		waitForLockWaiters(t, 1, served)
		timex.Advance(time.Second)
		if err := holder.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireBody(t, <-served, http.StatusForbidden, "")
	})

	checkVerifyGolden(t, api, journey+"/03_logs_and_registration_kept")
}
