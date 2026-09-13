package e2e_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestListActiveAnnouncementsWithoutAuthentication(t *testing.T) {
	api := newTestAPI(t)
	cutoff := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)

	for i, test := range []struct {
		name      string
		namespace string
	}{
		{name: "plain", namespace: "main"},
		{name: "escaped_slash", namespace: "a/b"},
		{name: "escaped_space", namespace: "hello world"},
		{name: "unicode", namespace: "日本語"},
	} {
		t.Run(test.name, func(t *testing.T) {
			id := fmt.Sprintf("11111111-1111-4111-8111-%012d", i+1)
			api.db.SeedAnnouncement(t, id, test.namespace, "Announcement", cutoff.Add(-time.Hour), cutoff.Add(time.Hour), false)

			timex.TheWorld(cutoff, func() {
				checkHTTPGolden(t, api.handler, test.name)
			})

			if api.proxied.Load() != 0 {
				t.Error("native read contacted an upstream")
			}
		})
	}
}

func TestActiveAnnouncementPublicationWindowAndLimit(t *testing.T) {
	api := newTestAPI(t)
	cutoff := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)

	for i, row := range []struct {
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
		id := fmt.Sprintf("22222222-2222-4222-8222-%012d", i+1)
		api.db.SeedAnnouncement(t, id, "main", row.title, row.start, row.end, row.deleted)
	}

	for i := 1; i <= 12; i++ {
		id := fmt.Sprintf("33333333-3333-4333-8333-%012d", i)
		title := fmt.Sprintf("older %02d", i)
		start := cutoff.Add(-time.Duration(i) * time.Minute)
		api.db.SeedAnnouncement(t, id, "main", title, start, cutoff.Add(time.Hour), false)
	}
	api.db.SeedAnnouncement(t, "44444444-4444-4444-8444-444444444444", "other", "other namespace", cutoff, cutoff.Add(time.Hour), false)

	timex.TheWorld(cutoff, func() {
		checkHTTPGolden(t, api.handler, "publication_window_and_limit")
	})
}

func TestListActiveAnnouncementsReturnsEmptyArray(t *testing.T) {
	t.Parallel()
	api := newTestAPI(t)

	checkHTTPGolden(t, api.handler, "empty")
}
