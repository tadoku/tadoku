package profile

import (
	"reflect"
	"sync"
	"testing"
	"time"

	kratosapi "github.com/ory/kratos-client-go"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testkratos"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestUserCacheUsesRealKratosAcrossPagesAndRetainsSnapshotOnFailure(t *testing.T) {
	fixture, err := testkratos.New(t.Context(), "testdata/kratos.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	cache := NewUserCache(fixture.CursorClient())
	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return instant }
	const readers = 8
	results := make(chan []cachedUser, readers)
	var ready sync.WaitGroup
	var done sync.WaitGroup
	start := make(chan struct{})
	ready.Add(readers)
	done.Add(readers)

	for range readers {
		go func() {
			defer done.Done()
			ready.Done()
			<-start
			users, err := cache.Users(t.Context())
			if err != nil {
				t.Error(err)
				return
			}
			results <- users
		}()
	}
	ready.Wait()
	close(start)
	done.Wait()
	close(results)

	for users := range results {
		if len(users) != 501 {
			t.Fatalf("users=%d, want identities from both cursor pages", len(users))
		}
		if users[0].ID != "10000000-0000-4000-8000-000000000001" || users[500].ID != "10000000-0000-4000-8000-000000000501" {
			t.Errorf("provider order bounds=%q..%q", users[0].ID, users[500].ID)
		}
		if users[0].CreatedAt != "2026-09-12T13:14:15Z" {
			t.Errorf("created_at=%q, want legacy formatting", users[0].CreatedAt)
		}
	}

	deletedID := "10000000-0000-4000-8000-000000000501"
	if _, err := fixture.Client().IdentityApi.DeleteIdentity(t.Context(), deletedID).Execute(); err != nil {
		t.Fatal(err)
	}

	cache.now = func() time.Time { return instant.Add(userCacheRefreshInterval - time.Second) }
	fresh, err := cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh) != 501 {
		t.Fatalf("fresh users=%d, want cached 501", len(fresh))
	}
	fresh[0].DisplayName = "mutated"
	second, err := cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if second[0].DisplayName == "mutated" {
		t.Error("Users returned the mutable cache snapshot")
	}

	cache.now = func() time.Time { return instant.Add(userCacheRefreshInterval) }
	refreshed, err := cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(refreshed) != 500 {
		t.Fatalf("refreshed users=%d, want provider deletion reflected", len(refreshed))
	}

	if err := fixture.Close(); err != nil {
		t.Fatal(err)
	}
	cache.now = func() time.Time { return instant.Add(2 * userCacheRefreshInterval) }
	stale, err := cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stale, refreshed) {
		t.Errorf("failed refresh snapshot=%+v, want %+v", stale, refreshed)
	}

	cold := NewUserCache(fixture.CursorClient())
	if _, err := cold.Users(t.Context()); err == nil {
		t.Fatal("cold cache accepted an unavailable identity provider")
	}
}

func TestUserCacheTTLUsesInjectedRealClockIndependentOfBusinessTime(t *testing.T) {
	fixture, err := testkratos.New(t.Context(), "testdata/kratos.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	cache := NewUserCache(fixture.CursorClient())
	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return instant }

	first, err := cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 501 {
		t.Fatalf("users=%d, want 501", len(first))
	}

	deletedID := "10000000-0000-4000-8000-000000000501"
	if _, err := fixture.Client().IdentityApi.DeleteIdentity(t.Context(), deletedID).Execute(); err != nil {
		t.Fatal(err)
	}

	timex.TheWorld(instant.Add(userCacheRefreshInterval+time.Minute), func() {
		cached, err := cache.Users(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if len(cached) != 501 {
			t.Fatalf("business-time travel users=%d, want cached 501", len(cached))
		}
	})

	cache.now = func() time.Time { return instant.Add(userCacheRefreshInterval) }
	refreshed, err := cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(refreshed) != 500 {
		t.Fatalf("real-time refresh users=%d, want provider deletion reflected", len(refreshed))
	}
}

func TestFromIdentityFiltersAndFormatsProviderIdentities(t *testing.T) {
	createdAt := time.Date(2026, 9, 12, 13, 14, 15, 0, time.FixedZone("test", 2*60*60))
	tests := []struct {
		name     string
		identity kratosapi.Identity
		want     cachedUser
		wantOK   bool
	}{
		{
			name: "user",
			identity: kratosapi.Identity{
				Id:        "user-id",
				SchemaId:  "user",
				Traits:    map[string]any{"display_name": "Reader", "email": "reader@example.test"},
				CreatedAt: &createdAt,
			},
			want: cachedUser{
				ID:          "user-id",
				DisplayName: "Reader",
				Email:       "reader@example.test",
				CreatedAt:   "2026-09-12T13:14:15Z",
			},
			wantOK: true,
		},
		{name: "other schema", identity: kratosapi.Identity{SchemaId: "service"}},
		{name: "invalid traits", identity: kratosapi.Identity{SchemaId: "user", Traits: "not an object"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := fromIdentity(test.identity)
			if ok != test.wantOK || !reflect.DeepEqual(got, test.want) {
				t.Errorf("fromIdentity()=(%+v, %t), want (%+v, %t)", got, ok, test.want, test.wantOK)
			}
		})
	}
}
