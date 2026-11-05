package e2e_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestContestEventsOfALogFollowCommitOrderAcrossClocks(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		createdAt time.Time
		removedAt time.Time
	}{
		{
			name:      "later_commit_reads_an_earlier_clock",
			createdAt: fixtureInstant.Add(time.Minute),
			removedAt: fixtureInstant,
		},
		{
			name:      "both_commits_read_the_same_clock",
			createdAt: fixtureInstant,
			removedAt: fixtureInstant,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const role = "clock"
			const contestID = "f0000000-0000-4000-8000-000000000002"
			s := newScenario(t, role)
			s.seed(t, role, tenantSeed{setup: []string{filepath.Join(journeysDir, "ContestEvents", "setup.sql")}})
			s.useIDs(descendingUUIDs())

			create := httptest.NewRequest(
				http.MethodPost,
				"/immersion/logs",
				strings.NewReader(`{"language_code":"eng","activity_id":1,"amount":8,"unit_key":"pages",`+
					`"registration_ids":["e0000000-0000-4000-8000-000000000002"]}`),
			)
			create.Header.Set("Content-Type", "application/json")
			s.clock.Set(tc.createdAt)
			created := s.serveWith(t, scoringEnabledHandler, role, user, create)
			requireBody(t, created, http.StatusOK, "")
			var log struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(created.Body.Bytes(), &log); err != nil {
				t.Fatal(err)
			}

			s.clock.Set(tc.removedAt)
			remove := httptest.NewRequest(http.MethodDelete, "/immersion/logs/"+log.ID, nil)
			requireBody(t, s.serveWith(t, scoringEnabledHandler, role, user, remove), http.StatusOK, "")

			read := httptest.NewRequest(http.MethodGet, "/immersion/contests/"+contestID+"/events", nil)
			response := s.serve(t, role, user, read)
			var page concurrentContestEventPage
			if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
				t.Fatalf("status=%d body=%s: %v", response.Code, response.Body, err)
			}

			var history []string
			for _, event := range page.Events {
				if event.LogID == log.ID {
					history = append(history, event.Event+"@"+event.OccurredAt.Format(time.StampMicro))
				}
			}
			if len(history) != 2 ||
				!strings.HasPrefix(history[0], "log.submitted@") ||
				!strings.HasPrefix(history[1], "log.withdrawn@") {
				t.Fatalf(
					"the deleted log's events in feed order are %v, want its submission and then its withdrawal",
					history,
				)
			}
		})
	}
}

// descendingUUIDs makes each new UUID sort before the previous one, so an equal-timestamp tie broken by event ID
// orders a later event first.
func descendingUUIDs() io.Reader {
	randomBytes := make([]byte, 16*128)
	for index := range 128 {
		randomBytes[index*16+15] = byte(128 - index)
	}
	return bytes.NewReader(randomBytes)
}
