package profile

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	kratosclient "github.com/tadoku/tadoku/services/tadoku-api/infra/kratos"
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

	if err := mapIdentityFetchError(kratosclient.ErrNotFound); !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("mapIdentityFetchError(ErrNotFound) = %v, want ErrIdentityNotFound", err)
	}

	wrapped := fmt.Errorf("fetch: %w", kratosclient.ErrNotFound)
	if err := mapIdentityFetchError(wrapped); !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("mapIdentityFetchError(wrapped ErrNotFound) = %v, want ErrIdentityNotFound", err)
	}

	other := errors.New("provider unavailable")
	if err := mapIdentityFetchError(other); !errors.Is(err, other) {
		t.Errorf("mapIdentityFetchError(other) = %v, want %v", err, other)
	}
}
