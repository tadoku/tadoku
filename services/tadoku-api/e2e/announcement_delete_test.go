package e2e_test

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

func TestDeleteAnnouncement(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusNoContent},
		{description: []string{"missing"}, want: http.StatusNoContent},
		{description: []string{"already", "deleted"}, want: http.StatusNoContent},
		{description: []string{"wrong", "namespace"}, want: http.StatusNoContent},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"active", "as", "id"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
	}

	for _, test := range tests {
		name := APITestName("DeleteAnnouncement", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler},
			)
		})
	}
}

func TestDeleteAnnouncementPreservesRow(t *testing.T) {
	deleteDir := filepath.Join("testdata", APITestName("DeleteAnnouncement", http.StatusNoContent, "admin"))
	wrongNamespaceDir := filepath.Join("testdata", APITestName("DeleteAnnouncement", http.StatusNoContent, "wrong", "namespace"))
	foundDir := filepath.Join("testdata", APITestName("FindAnnouncementByID", http.StatusOK, "admin"))
	missingDir := filepath.Join("testdata", APITestName("FindAnnouncementByID", http.StatusNotFound, "not", "found"))
	const snapshotSQL = `select (to_jsonb(announcements) - 'deleted_at')::text, deleted_at
		from announcements where id = '11111111-1111-4111-8111-111111111111'`

	for _, impl := range []implementation{
		{name: "tadoku-api", handler: api.handler},
		{name: "content-api", handler: legacyContent.handler},
	} {
		t.Run(impl.name, func(t *testing.T) {
			api.reset(t, deleteDir)
			var before string
			var initialDeletedAt *time.Time
			if err := api.db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&before, &initialDeletedAt); err != nil {
				t.Fatal(err)
			}
			if initialDeletedAt != nil {
				t.Fatal("seed announcement is already deleted")
			}

			atFixtureInstant(func() {
				checkHTTPGolden(t, impl.handler, wrongNamespaceDir, http.StatusNoContent)
				checkHTTPGolden(t, impl.handler, foundDir, http.StatusOK)
				checkHTTPGolden(t, impl.handler, deleteDir, http.StatusNoContent)
				checkHTTPGolden(t, impl.handler, missingDir, http.StatusNotFound)
			})

			var after string
			var deletedAt *time.Time
			if err := api.db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&after, &deletedAt); err != nil {
				t.Fatalf("soft-deleted row must still exist: %v", err)
			}
			if after != before {
				t.Errorf("deletion changed fields other than deleted_at:\nbefore %s\nafter %s", before, after)
			}
			if deletedAt == nil {
				t.Fatal("soft-deleted row has no deleted_at")
			}
			if impl.name == "tadoku-api" && !deletedAt.Equal(fixtureInstant) {
				t.Errorf("deleted_at=%v, want business time %v", deletedAt, fixtureInstant)
			}

			atFixtureInstant(func() { checkHTTPGolden(t, impl.handler, deleteDir, http.StatusNoContent) })
			var repeatedDeletedAt time.Time
			if err := api.db.Pool.QueryRow(t.Context(), "select deleted_at from announcements where id = '11111111-1111-4111-8111-111111111111'").Scan(&repeatedDeletedAt); err != nil {
				t.Fatal(err)
			}
			if !repeatedDeletedAt.Equal(*deletedAt) {
				t.Errorf("repeated deletion changed deleted_at from %v to %v", deletedAt, repeatedDeletedAt)
			}
			if api.proxied.Load() != 0 {
				t.Error("handler contacted an upstream")
			}
		})
	}
}
