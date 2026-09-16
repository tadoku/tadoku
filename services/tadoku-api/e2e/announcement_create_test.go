package e2e_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateAnnouncement(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"admin"}, want: http.StatusCreated},
		{description: []string{"null", "href"}, want: http.StatusCreated},
		{description: []string{"omitted", "href"}, want: http.StatusCreated},
		{description: []string{"empty", "href"}, want: http.StatusCreated},
		{description: []string{"ignored", "server", "fields"}, want: http.StatusCreated},
		{description: []string{"trailing", "json"}, want: http.StatusCreated},
		{description: []string{"xml"}, want: http.StatusCreated},
		{description: []string{"offset", "dates"}, want: http.StatusCreated},
		{description: []string{"empty", "title"}, want: http.StatusBadRequest},
		{description: []string{"empty", "content"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "style"}, want: http.StatusBadRequest},
		{description: []string{"missing", "start"}, want: http.StatusBadRequest},
		{description: []string{"missing", "end"}, want: http.StatusBadRequest},
		{description: []string{"equal", "dates"}, want: http.StatusBadRequest},
		{description: []string{"reversed", "dates"}, want: http.StatusBadRequest},
		{description: []string{"zero", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "time"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "href"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "guest"}, want: http.StatusBadRequest},
		{description: []string{"whitespace", "guest"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "xml"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"null", "body"}, want: http.StatusBadRequest},
		{description: []string{"missing", "content", "type"}, want: http.StatusBadRequest},
		{description: []string{"unsupported", "content", "type"}, want: http.StatusBadRequest},
		{description: []string{"form"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "form"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"empty", "guest"}, want: http.StatusUnauthorized},
		{description: []string{"null", "guest"}, want: http.StatusUnauthorized},
		{description: []string{"form", "guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin"}, want: http.StatusForbidden},
		{description: []string{"invalid", "non", "admin"}, want: http.StatusForbidden},
		{description: []string{"empty", "non", "admin"}, want: http.StatusForbidden},
		{description: []string{"duplicate", "id"}, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		name := APITestName("CreateAnnouncement", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{name: "content-api", handler: legacyContent.handler},
			)
		})
	}
}

func TestCreateAnnouncementCommitsForReadback(t *testing.T) {
	// Both pgx versions discard offsets for timestamp-without-time-zone storage;
	// the create response still echoes the input offsets unchanged.
	createDir := filepath.Join("testdata", APITestName("CreateAnnouncement", http.StatusCreated, "offset", "dates"))
	readDir := filepath.Join("testdata", APITestName("FindAnnouncementByID", http.StatusOK, "created"))
	for _, impl := range []implementation{
		{name: "tadoku-api", handler: api.handler},
		{name: "content-api", handler: legacyContent.handler},
	} {
		t.Run(impl.name, func(t *testing.T) {
			api.reset(t, createDir)
			atFixtureInstant(func() {
				checkHTTPGolden(t, impl.handler, createDir, http.StatusCreated)
				checkHTTPGolden(t, impl.handler, readDir, http.StatusOK)
			})
			if api.proxied.Load() != 0 {
				t.Error("handler contacted an upstream")
			}
		})
	}
}

func TestCreateAnnouncementGeneratesID(t *testing.T) {
	// UUID generation is deliberately checked separately from deterministic HTTP
	// goldens, without replacing or normalizing the generated response ID.
	dir := filepath.Join("testdata", APITestName("CreateAnnouncement", http.StatusCreated, "generated", "id"))
	for _, impl := range []implementation{
		{name: "tadoku-api", handler: api.handler},
		{name: "content-api", handler: legacyContent.handler},
	} {
		t.Run(impl.name, func(t *testing.T) {
			api.reset(t, dir)
			input, err := os.ReadFile(filepath.Join(dir, "request.http"))
			if err != nil {
				t.Fatal(err)
			}
			request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(input)))
			if err != nil {
				t.Fatal(err)
			}
			defer request.Body.Close()
			response := httptest.NewRecorder()
			atFixtureInstant(func() { impl.handler.ServeHTTP(response, request) })
			if response.Code != http.StatusCreated {
				t.Fatalf("status=%d, want 201: %s", response.Code, response.Body.String())
			}
			var body struct {
				ID        uuid.UUID `json:"id"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.ID == uuid.Nil || body.ID.Version() != 4 || body.ID.Variant() != uuid.RFC4122 {
				t.Errorf("generated UUID=%s, want nonzero v4 RFC4122 UUID", body.ID)
			}
			if !body.CreatedAt.Equal(fixtureInstant) || !body.UpdatedAt.Equal(fixtureInstant) {
				t.Errorf("response timestamps=%s, %s, want %s", body.CreatedAt, body.UpdatedAt, fixtureInstant)
			}
			var createdAt, updatedAt time.Time
			if err := api.db.Pool.QueryRow(t.Context(), "select created_at, updated_at from announcements where id = $1 and namespace = 'main'", body.ID).Scan(&createdAt, &updatedAt); err != nil {
				t.Fatal(err)
			}
			if !createdAt.Equal(fixtureInstant) || !updatedAt.Equal(fixtureInstant) {
				t.Errorf("persisted timestamps=%s, %s, want %s", createdAt, updatedAt, fixtureInstant)
			}
			if api.proxied.Load() != 0 {
				t.Error("handler contacted an upstream")
			}
		})
	}
}
