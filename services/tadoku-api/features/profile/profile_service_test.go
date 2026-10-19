package profile

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestUserPagePreservesProviderOrder(t *testing.T) {
	users := make([]cachedUser, 105)
	for i := range users {
		users[i] = cachedUser{ID: fmt.Sprintf("user-%03d", i), DisplayName: fmt.Sprintf("User %03d", i)}
	}

	for _, test := range []struct {
		name      string
		pageSize  int
		page      int
		wantCount int
		wantFirst string
		wantLast  string
	}{
		{name: "default page size", pageSize: 0, wantCount: 20, wantFirst: "user-000", wantLast: "user-019"},
		{name: "negative page starts at zero", pageSize: 2, page: -4, wantCount: 2, wantFirst: "user-000", wantLast: "user-001"},
		{name: "page size is capped", pageSize: 1000, wantCount: 100, wantFirst: "user-000", wantLast: "user-099"},
		{name: "later page", pageSize: 20, page: 5, wantCount: 5, wantFirst: "user-100", wantLast: "user-104"},
		{name: "empty page", pageSize: 20, page: 6, wantCount: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			pageSize, page := normalizeUserPage(test.pageSize, test.page)
			got := userPage(users, pageSize, page)
			if len(got) != test.wantCount {
				t.Fatalf("page size = %d, want %d", len(got), test.wantCount)
			}
			if test.wantCount > 0 && (got[0].ID != test.wantFirst || got[len(got)-1].ID != test.wantLast) {
				t.Errorf("page bounds = %q..%q, want %q..%q", got[0].ID, got[len(got)-1].ID, test.wantFirst, test.wantLast)
			}
		})
	}
}

func TestSearchUsersRanksCaseInsensitiveMatches(t *testing.T) {
	users := []cachedUser{
		{ID: "bobby", DisplayName: "Bobby", Email: "first@example.test"},
		{ID: "bob", DisplayName: "Bob", Email: "second@example.test"},
		{ID: "bob-user", DisplayName: "Bob User", Email: "user@example.test"},
		{ID: "alice", DisplayName: "Alice", Email: "alice@example.test"},
	}

	matches := searchUsers(users, "BOB")
	if len(matches) != 3 {
		t.Fatalf("search total = %d, want 3", len(matches))
	}
	want := []cachedUser{users[1], users[0], users[2]}
	if !reflect.DeepEqual(matches, want) {
		t.Errorf("matches = %+v, want %+v", matches, want)
	}
}

func TestDecodeIdentityTraitsRoundTripsDisplayNameWithoutEmail(t *testing.T) {
	traits, err := decodeIdentityTraits(map[string]any{"display_name": "Reader"})
	if err != nil {
		t.Fatal(err)
	}
	if traits.DisplayName != "Reader" {
		t.Errorf("DisplayName = %q, want Reader", traits.DisplayName)
	}
	if traits.Email != "" {
		t.Errorf("Email = %q, want empty when absent from traits", traits.Email)
	}
}

func TestMapIdentityFetchErrorMapsKratosNotFound(t *testing.T) {
	t.Parallel()

	if err := mapIdentityFetchError(identity.ErrLegacyNotFound); !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("mapIdentityFetchError(ErrNotFound) = %v, want ErrIdentityNotFound", err)
	}

	wrapped := fmt.Errorf("fetch: %w", identity.ErrLegacyNotFound)
	if err := mapIdentityFetchError(wrapped); !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("mapIdentityFetchError(wrapped ErrNotFound) = %v, want ErrIdentityNotFound", err)
	}

	other := errors.New("provider unavailable")
	if err := mapIdentityFetchError(other); errx.KindOf(err) != errx.Unavailable || !errors.Is(err, other) {
		t.Errorf("mapIdentityFetchError(other) = %v, want Unavailable wrapping %v", err, other)
	}
}

func TestFindProfileAndFetchAccountCreatedAtTreatNonUserSchemaAsNotFound(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/admin/identities/"+id.String() {
			t.Errorf("request=%s %s", r.Method, r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"` + id.String() + `",
			"schema_id":"service",
			"schema_url":"http://kratos/schemas/service",
			"state":"active",
			"traits":{}
		}`))
	}))
	t.Cleanup(provider.Close)

	httpClient := provider.Client()
	httpClient.Timeout = time.Second
	t.Cleanup(httpClient.CloseIdleConnections)
	service := NewService(
		nil,
		nil,
		nil,
		identity.NewLegacyReader(provider.URL, identity.WithHTTPClient(httpClient)),
	)

	if _, err := service.FindProfile(t.Context(), id); !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("FindProfile() error = %v, want ErrIdentityNotFound", err)
	}
	if _, err := service.FetchAccountCreatedAt(t.Context(), id); !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("FetchAccountCreatedAt() error = %v, want ErrIdentityNotFound", err)
	}
}
