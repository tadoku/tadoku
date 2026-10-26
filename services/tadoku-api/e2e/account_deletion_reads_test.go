package e2e_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type leaderboardEntry struct {
	UserID          string  `json:"user_id"`
	UserDisplayName string  `json:"user_display_name"`
	Score           float32 `json:"score"`
}

func TestAccountDeletionUpdatesLeaderboardsAndContestReads(t *testing.T) {
	resetAccountDeletion(t)
	const finished = "/immersion/contests/f0000000-0000-4000-8000-0000000000a1/leaderboard"
	unfinished := []string{
		"/immersion/contests/f0000000-0000-4000-8000-0000000000a2/leaderboard",
		"/immersion/leaderboard/yearly/2026",
		"/immersion/leaderboard/global",
	}
	before := leaderboardEntries(t, finished)[bannedUserID]
	for _, path := range unfinished {
		if _, listed := leaderboardEntries(t, path)[bannedUserID]; !listed {
			t.Fatalf("%s does not list the user before the deletion", path)
		}
	}

	if _, err := fileAdminDeletion(t, bannedUserID, adminUserID, "Owner confirmed by email"); err != nil {
		t.Fatal(err)
	}
	runDeletionWorker(t)

	for _, path := range unfinished {
		if entry, listed := leaderboardEntries(t, path)[bannedUserID]; listed {
			t.Errorf("%s still lists the deleted user: %+v", path, entry)
		}
	}
	after := leaderboardEntries(t, finished)[bannedUserID]
	if after.UserDisplayName != "Deleted participant" || after.Score != before.Score || before.Score == 0 {
		t.Errorf("finished contest entry before=%+v after=%+v, want the same score as Deleted participant", before, after)
	}

	var ongoing struct {
		Registrations []struct {
			Contest struct {
				ID                   string `json:"id"`
				OwnerUserDisplayName string `json:"owner_user_display_name"`
			} `json:"contest"`
		} `json:"registrations"`
	}
	readAs(t, user2, "/immersion/contests/ongoing-registrations", &ongoing)
	organizers := make(map[string]string)
	for _, registration := range ongoing.Registrations {
		organizers[registration.Contest.ID] = registration.Contest.OwnerUserDisplayName
	}
	if organizer := organizers["f0000000-0000-4000-8000-0000000000a3"]; organizer != "Deleted organizer" {
		t.Errorf("ongoing registration organizer=%q, want Deleted organizer", organizer)
	}
}

func leaderboardEntries(t *testing.T, path string) map[string]leaderboardEntry {
	t.Helper()
	var page struct {
		Entries []leaderboardEntry `json:"entries"`
	}
	readAs(t, guest, path, &page)

	entries := make(map[string]leaderboardEntry, len(page.Entries))
	for _, entry := range page.Entries {
		entries[entry.UserID] = entry
	}
	return entries
}

func readAs(t *testing.T, who member, path string, target any) {
	t.Helper()
	tokens, err := loadCast(nil)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", path, nil)
	authorize(request, who, tokens)

	response := httptest.NewRecorder()
	atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })
	if response.Code != http.StatusOK {
		t.Fatalf("GET %s: status=%d body=%s", path, response.Code, response.Body)
	}
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
}
