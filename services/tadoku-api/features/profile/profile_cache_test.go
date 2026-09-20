package profile

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	kratosapi "github.com/ory/kratos-client-go"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type identityProviderFunc func(context.Context, int64, string) ([]kratosapi.Identity, string, error)

func (f identityProviderFunc) ListIdentities(ctx context.Context, pageSize int64, pageToken string) ([]kratosapi.Identity, string, error) {
	return f(ctx, pageSize, pageToken)
}

type suppressionRepositoryFunc func(context.Context) ([]string, error)

func (f suppressionRepositoryFunc) ListAccountDeletionSuppressedIdentityIDs(ctx context.Context) ([]string, error) {
	return f(ctx)
}

func kratosIdentity(id, schemaID, name, email string, createdAt *time.Time) kratosapi.Identity {
	return kratosapi.Identity{
		Id:        id,
		SchemaId:  schemaID,
		Traits:    map[string]any{"display_name": name, "email": email},
		CreatedAt: createdAt,
	}
}

func TestUserCacheRefreshUsesAllCursorPagesAndFiltersBeforeDeduplication(t *testing.T) {
	createdAt := time.Date(2026, 9, 12, 13, 14, 15, 0, time.FixedZone("test", 2*60*60))
	ctx := t.Context()
	requests := make([]string, 0, 2)
	provider := identityProviderFunc(func(gotCtx context.Context, pageSize int64, pageToken string) ([]kratosapi.Identity, string, error) {
		if gotCtx != ctx {
			t.Error("identity provider did not receive request context")
		}
		if pageSize != 500 {
			t.Errorf("page size=%d, want 500", pageSize)
		}
		requests = append(requests, pageToken)
		switch pageToken {
		case "":
			invalid := kratosIdentity("later-valid", "user", "ignored", "ignored@example.test", nil)
			invalid.Traits = "not an object"
			return []kratosapi.Identity{
				kratosIdentity("non-user", "service", "Service", "service@example.test", nil),
				invalid,
				kratosIdentity("first", "user", "First", "first@example.test", &createdAt),
				kratosIdentity("first", "user", "Duplicate", "duplicate@example.test", nil),
			}, "next", nil
		case "next":
			return []kratosapi.Identity{
				kratosIdentity("later-valid", "user", "Later", "later@example.test", nil),
			}, "", nil
		default:
			return nil, "", errors.New("unexpected page token")
		}
	})
	cache := NewUserCache(provider, suppressionRepositoryFunc(func(gotCtx context.Context) ([]string, error) {
		if gotCtx != ctx {
			t.Error("suppression repository did not receive request context")
		}
		return nil, nil
	}))

	got, err := cache.Users(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(requests, []string{"", "next"}) {
		t.Fatalf("page tokens=%v, want initial and next", requests)
	}
	want := []CachedUser{
		{ID: "first", DisplayName: "First", Email: "first@example.test", CreatedAt: "2026-09-12T13:14:15Z"},
		{ID: "later-valid", DisplayName: "Later", Email: "later@example.test"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("users=%+v, want %+v", got, want)
	}

	got[0].DisplayName = "mutated"
	got, err = cache.Users(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].DisplayName != "First" {
		t.Error("Users returned the mutable cache snapshot")
	}
}

func TestUserCacheSerializesColdLoadAndCachesEmptySnapshot(t *testing.T) {
	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	var calls atomic.Int32
	cache := NewUserCache(identityProviderFunc(func(context.Context, int64, string) ([]kratosapi.Identity, string, error) {
		calls.Add(1)
		return []kratosapi.Identity{}, "", nil
	}), suppressionRepositoryFunc(func(context.Context) ([]string, error) {
		return nil, nil
	}))

	timex.TheWorld(instant, func() {
		const readers = 16
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
				}
				if len(users) != 0 {
					t.Errorf("users=%+v, want empty", users)
				}
			}()
		}
		ready.Wait()
		close(start)
		done.Wait()
	})
	if got := calls.Load(); got != 1 {
		t.Fatalf("cold load calls=%d, want 1", got)
	}

	timex.TheWorld(instant.Add(userCacheRefreshInterval-time.Second), func() {
		if _, err := cache.Users(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
	if got := calls.Load(); got != 1 {
		t.Errorf("fresh cache calls=%d, want 1", got)
	}

	timex.TheWorld(instant.Add(userCacheRefreshInterval), func() {
		if _, err := cache.Users(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
	if got := calls.Load(); got != 2 {
		t.Errorf("expired cache calls=%d, want 2", got)
	}
}

func TestUserCacheRefreshFailureAndStickySuppressionPolicy(t *testing.T) {
	providerErr := errors.New("kratos unavailable")
	suppressionErr := errors.New("postgres unavailable")
	providerMode := "failed"
	provider := identityProviderFunc(func(_ context.Context, _ int64, pageToken string) ([]kratosapi.Identity, string, error) {
		switch providerMode {
		case "failed":
			return nil, "", providerErr
		case "later-failed":
			if pageToken == "" {
				return []kratosapi.Identity{kratosIdentity("partial", "user", "Partial", "partial@example.test", nil)}, "next", nil
			}
			return nil, "", providerErr
		case "repeated":
			return []kratosapi.Identity{kratosIdentity("partial", "user", "Partial", "partial@example.test", nil)}, "repeat", nil
		default:
			return []kratosapi.Identity{
				kratosIdentity("suppressed", "user", "Suppressed", "suppressed@example.test", nil),
				kratosIdentity("visible", "user", "Visible", "visible@example.test", nil),
			}, "", nil
		}
	})
	suppressionMode := "present"
	suppressions := suppressionRepositoryFunc(func(context.Context) ([]string, error) {
		switch suppressionMode {
		case "failed":
			return nil, suppressionErr
		case "absent":
			return nil, nil
		default:
			return []string{"suppressed"}, nil
		}
	})
	cache := NewUserCache(provider, suppressions)
	refresh := func() error {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		return cache.refresh(t.Context(), timex.Now())
	}

	if _, err := cache.Users(t.Context()); !errors.Is(err, providerErr) {
		t.Fatalf("cold provider error=%v, want Kratos failure", err)
	}
	providerMode = "ok"
	got, err := cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "visible" {
		t.Fatalf("initial users=%+v, want visible only", got)
	}

	providerMode = "later-failed"
	if err := refresh(); !errors.Is(err, providerErr) {
		t.Fatalf("provider error=%v, want Kratos failure", err)
	}
	got, err = cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "visible" {
		t.Errorf("provider failure replaced snapshot: %+v", got)
	}

	providerMode = "ok"
	suppressionMode = "failed"
	if err := refresh(); !errors.Is(err, suppressionErr) {
		t.Fatalf("suppression error=%v, want PostgreSQL failure", err)
	}
	got, err = cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("suppression failure retained users: %+v", got)
	}

	suppressionMode = "absent"
	if err := refresh(); err != nil {
		t.Fatal(err)
	}
	got, err = cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "visible" {
		t.Errorf("sticky suppression after successful refresh=%+v, want visible only", got)
	}

	providerMode = "repeated"
	if err := refresh(); err == nil {
		t.Fatal("repeated continuation token was accepted")
	}
	got, err = cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "visible" {
		t.Errorf("repeated token replaced complete snapshot: %+v", got)
	}

	providerMode = "ok"
	suppressionMode = "failed"
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	cache.mu.Lock()
	cache.checkedAt = time.Time{}
	err = cache.refresh(canceled, timex.Now())
	cleared := len(cache.users) == 0
	cache.mu.Unlock()
	if !errors.Is(err, suppressionErr) {
		t.Errorf("canceled suppression error=%v, want repository failure", err)
	}
	if !cleared {
		t.Error("canceled suppression query retained visible users")
	}

	suppressionMode = "absent"
	got, err = cache.Users(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "visible" {
		t.Errorf("retry after canceled refresh=%+v, want visible only", got)
	}
}
