package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestContestEventConcurrentPublication(t *testing.T) {
	previous := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time { return fixtureInstant }
	defer func() { jwt.TimeFunc = previous }()

	for _, commitPending := range []bool{true, false} {
		t.Run(fmt.Sprintf("commit_pending_%t", commitPending), func(t *testing.T) {
			cleanupContestEventUnit(t)
			resetJourney(t, api, "testdata/journeys/ContestEventConcurrentPublication")
			tokens, err := loadCast([]step{{as: guest}, {as: user}, {as: user2}})
			if err != nil {
				t.Fatal(err)
			}

			proveContestEventPublication(t, tokens, commitPending)
		})
	}
}

func proveContestEventPublication(t *testing.T, tokens map[member]string, commitPending bool) {
	t.Helper()
	const contestID = "f0000000-0000-4000-8000-000000000001"
	const eventID = "90000000-0000-4000-8000-000000000001"
	const logID = "d0000000-0000-4000-8000-000000000003"
	const feed = "/immersion/contests/" + contestID + "/events"
	writtenAt := fixtureInstant.Add(123456 * time.Microsecond)
	pendingAt := writtenAt.Add(-time.Minute)

	tx, err := api.db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(t.Context(), `update contests
		set title = 'Pending contest edit'
		where tenant = 'tadoku/prod'
			and id = $1`, contestID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(t.Context(), `insert into contest_events (tenant, id, contest_id, log_id, event, occurred_at)
		values ('tadoku/prod', $1, $2, 'd0000000-0000-4000-8000-000000000001', 'log.updated', $3)`,
		eventID, contestID, pendingAt)
	if err != nil {
		t.Fatal(err)
	}

	for _, suffix := range []string{"", "?cursor=now"} {
		page := readConcurrentContestEvents(t, feed+suffix, tokens)
		if len(page.Events) != 0 || page.NextCursor != "0" || page.HasMore {
			t.Fatalf("pending transaction leaked a checkpoint: %+v", page)
		}
	}

	publish := func(at time.Time, request *http.Request) *httptest.ResponseRecorder {
		t.Helper()
		defer request.Body.Close()
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		response := httptest.NewRecorder()
		timex.TheWorld(at, func() {
			scoringEnabledHandler.ServeHTTP(response, request.WithContext(ctx))
		})
		if response.Code != http.StatusOK {
			t.Fatalf("independent write failed while the earlier transaction was open: status=%d body=%s",
				response.Code, response.Body)
		}
		return response
	}

	request := httptest.NewRequest(http.MethodPut, "/immersion/logs/"+logID,
		strings.NewReader(`{"amount":7,"unit_key":"reading_page","description":"concurrent update"}`))
	request.Header.Set("Content-Type", "application/json")
	authorize(request, user2, tokens)
	publish(writtenAt, request)

	updated := readConcurrentContestEvents(t, feed+"?limit=1", tokens)
	if len(updated.Events) != 1 || updated.Events[0].LogID != logID ||
		updated.Events[0].Event != "log.updated" || !updated.Events[0].OccurredAt.Equal(writtenAt) ||
		updated.NextCursor != updated.Events[0].ID || updated.HasMore {
		t.Fatalf("independent update is not visible at its source time: %+v", updated)
	}
	var sourceUpdatedAt time.Time
	err = api.db.Pool.QueryRow(t.Context(), `select updated_at
		from logs
		where tenant = 'tadoku/prod'
			and id = $1`, logID).Scan(&sourceUpdatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Events[0].OccurredAt.Equal(sourceUpdatedAt) {
		t.Fatalf("event occurred_at=%s differs from source updated_at=%s",
			updated.Events[0].OccurredAt, sourceUpdatedAt)
	}

	createdAt := writtenAt.Add(time.Minute)
	create := readHTTPRequest(t, "testdata/journeys/ContestEvents/11_create_and_submit_b")
	authorize(create, user, tokens)
	createdResponse := publish(createdAt, create)
	var created struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
	}
	if err := json.Unmarshal(createdResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if !created.CreatedAt.Equal(createdAt) {
		t.Fatalf("source created_at=%s, want %s", created.CreatedAt, createdAt)
	}

	committed := readConcurrentContestEvents(t, feed, tokens)
	if len(committed.Events) != 2 || committed.Events[0] != updated.Events[0] ||
		committed.Events[1].LogID != created.ID || committed.Events[1].Event != "log.submitted" ||
		!committed.Events[1].OccurredAt.Equal(created.CreatedAt) ||
		committed.NextCursor != committed.Events[1].ID || committed.HasMore {
		t.Fatalf("committed history leaked the pending event or changed source timestamps: %+v", committed)
	}
	now := readConcurrentContestEvents(t, feed+"?cursor=now", tokens)
	if len(now.Events) != 0 || now.NextCursor != committed.NextCursor || now.HasMore {
		t.Fatalf("now did not checkpoint the latest committed timestamp: %+v", now)
	}

	independent := httptest.NewRequest(http.MethodPut,
		"/immersion/logs/d0000000-0000-4000-8000-000000000004",
		strings.NewReader(`{"amount":5,"unit_key":"reading_page"}`))
	independent.Header.Set("Content-Type", "application/json")
	authorize(independent, user, tokens)
	publish(createdAt, independent)

	if commitPending {
		err = tx.Commit(t.Context())
	} else {
		err = tx.Rollback(t.Context())
	}
	if err != nil {
		t.Fatal(err)
	}

	timex.TheWorld(createdAt.Add(time.Hour), func() {
		history := readConcurrentContestEvents(t, feed, tokens)
		if commitPending {
			if len(history.Events) != 3 || history.Events[0].ID != eventID ||
				!history.Events[0].OccurredAt.Equal(pendingAt) {
				t.Fatalf("late commit did not retain its earlier timestamp: %+v", history)
			}
			history.Events = history.Events[1:]
		}
		if !slices.Equal(history.Events, committed.Events) ||
			history.NextCursor != committed.NextCursor || history.HasMore {
			t.Fatalf("committed events changed or a rolled-back event survived: %+v", history)
		}

		replayed := readConcurrentContestEvents(t, feed+"?cursor="+updated.NextCursor, tokens)
		if !slices.Equal(replayed.Events, committed.Events[1:]) ||
			replayed.NextCursor != committed.NextCursor || replayed.HasMore {
			t.Fatalf("replayed page changed its event timestamp: %+v", replayed)
		}
		tail := readConcurrentContestEvents(t, feed+"?cursor="+now.NextCursor, tokens)
		if len(tail.Events) != 0 || tail.NextCursor != now.NextCursor || tail.HasMore {
			t.Fatalf("late transaction changed the consumed checkpoint: %+v", tail)
		}
		nowAfter := readConcurrentContestEvents(t, feed+"?cursor=now", tokens)
		if len(nowAfter.Events) != 0 || nowAfter.NextCursor != now.NextCursor || nowAfter.HasMore {
			t.Fatalf("now followed commit order instead of event timestamps: %+v", nowAfter)
		}
	})
}

type concurrentContestEventPage struct {
	Events []struct {
		ID         string    `json:"id"`
		LogID      string    `json:"log_id"`
		Event      string    `json:"event"`
		OccurredAt time.Time `json:"occurred_at"`
	} `json:"events"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

func readConcurrentContestEvents(t *testing.T, path string, tokens map[member]string) concurrentContestEventPage {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	authorize(request, guest, tokens)
	response := httptest.NewRecorder()
	api.handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("event read failed: status=%d body=%s", response.Code, response.Body)
	}

	var page concurrentContestEventPage
	if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	return page
}
