package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestListActiveAnnouncementsWithoutAuthentication(t *testing.T) {
	f := newFixture(t)
	cutoff := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	start := cutoff.Add(-time.Hour)
	end := cutoff.Add(time.Hour)

	for _, test := range []struct {
		name      string
		namespace string
		path      string
	}{
		{name: "plain", namespace: "main", path: "main"},
		{name: "escaped slash", namespace: "a/b", path: "a%2Fb"},
		{name: "escaped space", namespace: "hello world", path: "hello%20world"},
		{name: "unicode", namespace: "日本語", path: "%E6%97%A5%E6%9C%AC%E8%AA%9E"},
	} {
		t.Run(test.name, func(t *testing.T) {
			seedID := f.db.SeedAnnouncement(t, test.namespace, "Announcement", start, end, false)
			id := uuid.MustParse(seedID)

			request := httptest.NewRequest(http.MethodGet, "/content/announcements/"+test.path+"/active", nil)
			response := httptest.NewRecorder()

			timex.TheWorld(cutoff, func() {
				f.handler.ServeHTTP(response, request)
			})

			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body)
			}
			if contentType := response.Header().Get("Content-Type"); contentType != "application/json; charset=UTF-8" {
				t.Errorf("Content-Type=%q", contentType)
			}

			var got openapi.ContentAnnouncements
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode announcements: %v", err)
			}

			want := openapi.ContentAnnouncements{
				Announcements: []openapi.ContentAnnouncement{{
					Id:        &id,
					Namespace: &test.namespace,
					Title:     "Announcement",
					Content:   "<p>Announcement</p>",
					Style:     openapi.ContentAnnouncementStyle("info"),
					Href:      nil,
					StartsAt:  start,
					EndsAt:    end,
					CreatedAt: &start,
					UpdatedAt: &start,
				}},
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("response=%+v, want %+v", got, want)
			}
			if f.proxied.Load() != 0 {
				t.Error("native read contacted an upstream")
			}
		})
	}
}

func TestActiveAnnouncementPublicationWindowAndLimit(t *testing.T) {
	f := newFixture(t)
	cutoff := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)

	for _, row := range []struct {
		title   string
		start   time.Time
		end     time.Time
		deleted bool
	}{
		{title: "starts now", start: cutoff, end: cutoff.Add(time.Hour)},
		{title: "ends now", start: cutoff.Add(-time.Hour), end: cutoff},
		{title: "future", start: cutoff.Add(time.Microsecond), end: cutoff.Add(time.Hour)},
		{title: "deleted", start: cutoff.Add(-time.Minute), end: cutoff.Add(time.Hour), deleted: true},
	} {
		f.db.SeedAnnouncement(t, "main", row.title, row.start, row.end, row.deleted)
	}

	for i := 1; i <= 12; i++ {
		title := fmt.Sprintf("older %02d", i)
		start := cutoff.Add(-time.Duration(i) * time.Minute)
		f.db.SeedAnnouncement(t, "main", title, start, cutoff.Add(time.Hour), false)
	}
	f.db.SeedAnnouncement(t, "other", "other namespace", cutoff, cutoff.Add(time.Hour), false)

	request := httptest.NewRequest(http.MethodGet, "/content/announcements/main/active", nil)
	response := httptest.NewRecorder()

	timex.TheWorld(cutoff, func() {
		f.handler.ServeHTTP(response, request)
	})

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}

	var got openapi.ContentAnnouncements
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode announcements: %v", err)
	}
	if len(got.Announcements) != 10 {
		t.Fatalf("got %d announcements, want 10", len(got.Announcements))
	}

	for i, item := range got.Announcements {
		want := "starts now"
		if i > 0 {
			want = fmt.Sprintf("older %02d", i)
		}
		if item.Title != want {
			t.Errorf("item %d=%q, want %q", i, item.Title, want)
		}
	}
}

func TestListActiveAnnouncementsReturnsEmptyArray(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	request := httptest.NewRequest(http.MethodGet, "/content/announcements/empty/active", nil)
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
	if response.Body.String() != "{\"announcements\":[]}\n" {
		t.Errorf("empty response=%s", response.Body)
	}
}
