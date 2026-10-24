package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestContestEventConcurrentPublication(t *testing.T) {
	for _, commitFirst := range []bool{true, false} {
		t.Run(fmt.Sprintf("commit_first_%t", commitFirst), func(t *testing.T) {
			cleanupContestEventUnit(t)
			resetJourney(t, api, "testdata/journeys/ContestEventConcurrentPublication")
			tokens, err := loadCast([]step{{as: guest}, {as: user}, {as: user2}})
			if err != nil {
				t.Fatal(err)
			}

			atFixtureInstant(func() {
				proveContestEventPublication(t, tokens, commitFirst)
			})
		})
	}
}

func proveContestEventPublication(t *testing.T, tokens map[member]string, commitFirst bool) {
	t.Helper()
	const contestID = "f0000000-0000-4000-8000-000000000001"
	const eventID = "90000000-0000-4000-8000-000000000001"
	const logID = "d0000000-0000-4000-8000-000000000003"
	const feed = "/immersion/contests/" + contestID + "/events"

	tx, err := api.db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())

	var backendPID int
	if err := tx.QueryRow(t.Context(), "select pg_backend_pid()").Scan(&backendPID); err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(t.Context(), `select id from contests
		where tenant = 'tadoku/prod' and id = $1 for no key update`, contestID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(t.Context(), `insert into contest_events (tenant, id, contest_id, log_id, event)
		values ('tadoku/prod', $1, $2, 'd0000000-0000-4000-8000-000000000001', 'log.updated')`,
		eventID, contestID)
	if err != nil {
		t.Fatal(err)
	}

	requestCtx, cancel := context.WithCancel(t.Context())
	defer cancel()
	request := httptest.NewRequest(http.MethodPut, "/immersion/logs/"+logID,
		strings.NewReader(`{"amount":7,"unit_key":"reading_page","description":"concurrent update"}`))
	request.Header.Set("Content-Type", "application/json")
	authorize(request, user2, tokens)
	request = request.WithContext(requestCtx)
	response := httptest.NewRecorder()
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		scoringEnabledHandler.ServeHTTP(response, request)
	}()
	defer func() {
		_ = tx.Rollback(context.Background())
		cancel()
		select {
		case <-finished:
		case <-time.After(5 * time.Second):
			t.Error("concurrent request did not finish after cleanup")
		}
	}()

	deadline := time.Now().Add(5 * time.Second)
	waiting := false
	for time.Now().Before(deadline) {
		select {
		case <-finished:
			t.Fatalf("second write completed before the earlier publication: status=%d body=%s",
				response.Code, response.Body)
		default:
		}
		err = api.db.Pool.QueryRow(t.Context(), `select exists (
			select 1 from pg_stat_activity
			where datname = current_database()
				and $1::integer = any(pg_blocking_pids(pid))
				and query like '%for no key update%'
		)`, backendPID).Scan(&waiting)
		if err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !waiting {
		t.Fatal("second write never reached the contest publication lock")
	}

	for _, suffix := range []string{"", "?cursor=now"} {
		page := readConcurrentContestEvents(t, feed+suffix, tokens)
		if len(page.Events) != 0 || page.NextCursor != "0" || page.HasMore {
			t.Fatalf("pending transaction leaked a checkpoint: %+v", page)
		}
	}

	independent := httptest.NewRequest(http.MethodPut,
		"/immersion/logs/d0000000-0000-4000-8000-000000000004",
		strings.NewReader(`{"amount":5,"unit_key":"reading_page"}`))
	independent.Header.Set("Content-Type", "application/json")
	authorize(independent, user, tokens)
	otherResponse := httptest.NewRecorder()
	scoringEnabledHandler.ServeHTTP(otherResponse, independent)
	if otherResponse.Code != http.StatusOK {
		t.Fatalf("unrelated contest write failed: status=%d body=%s", otherResponse.Code, otherResponse.Body)
	}

	if commitFirst {
		err = tx.Commit(t.Context())
	} else {
		err = tx.Rollback(t.Context())
	}
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("second write did not resume after the publication lock was released")
	}
	if response.Code != http.StatusOK {
		t.Fatalf("second write failed: status=%d body=%s", response.Code, response.Body)
	}

	first := readConcurrentContestEvents(t, feed+"?limit=1", tokens)
	if len(first.Events) != 1 || first.NextCursor != first.Events[0].ID {
		t.Fatalf("first committed page=%+v", first)
	}
	if commitFirst {
		if first.Events[0].ID != eventID || !first.HasMore {
			t.Fatalf("later commit overtook the first event: %+v", first)
		}
		second := readConcurrentContestEvents(t, feed+"?cursor="+first.NextCursor, tokens)
		if len(second.Events) != 1 || second.Events[0].LogID != logID ||
			second.Events[0].Event != "log.updated" || second.NextCursor != second.Events[0].ID || second.HasMore {
			t.Fatalf("second committed page=%+v", second)
		}
		first = second
	} else if first.Events[0].LogID != logID || first.Events[0].Event != "log.updated" || first.HasMore {
		t.Fatalf("rolled-back publication survived: %+v", first)
	}

	tail := readConcurrentContestEvents(t, feed+"?cursor="+first.NextCursor, tokens)
	if len(tail.Events) != 0 || tail.NextCursor != first.NextCursor || tail.HasMore {
		t.Fatalf("empty poll lost its checkpoint: %+v", tail)
	}
}

type concurrentContestEventPage struct {
	Events []struct {
		ID    string `json:"id"`
		LogID string `json:"log_id"`
		Event string `json:"event"`
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
