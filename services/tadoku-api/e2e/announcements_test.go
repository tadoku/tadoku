package e2e_test

import "testing"

func TestListActiveAnnouncementsWithoutAuthentication(t *testing.T) {
	for _, name := range []string{"plain", "escaped_slash", "escaped_space", "unicode"} {
		t.Run(name, func(t *testing.T) {
			reset(t, "testdata/announcements/namespaces.sql")
			checkHTTPGolden(t, api.handler, name)

			if api.proxied.Load() != 0 {
				t.Error("native read contacted an upstream")
			}
		})
	}
}

func TestActiveAnnouncementPublicationWindowAndLimit(t *testing.T) {
	reset(t, "testdata/announcements/publication_window_and_limit.sql")
	checkHTTPGolden(t, api.handler, "publication_window_and_limit")
}

func TestListActiveAnnouncementsReturnsEmptyArray(t *testing.T) {
	reset(t)
	checkHTTPGolden(t, api.handler, "empty")
}
