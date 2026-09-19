package profile

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	kratosapi "github.com/ory/kratos-client-go"
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
	requests := make([]string, 0, 2)
	provider := identityProviderFunc(func(_ context.Context, pageSize int64, pageToken string) ([]kratosapi.Identity, string, error) {
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
	cache := NewUserCache(provider, suppressionRepositoryFunc(func(context.Context) ([]string, error) {
		return nil, nil
	}))

	if err := cache.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(requests, []string{"", "next"}) {
		t.Fatalf("page tokens=%v, want initial and next", requests)
	}
	want := []CachedUser{
		{ID: "first", DisplayName: "First", Email: "first@example.test", CreatedAt: "2026-09-12T13:14:15Z"},
		{ID: "later-valid", DisplayName: "Later", Email: "later@example.test"},
	}
	if got := cache.Users(); !reflect.DeepEqual(got, want) {
		t.Errorf("users=%+v, want %+v", got, want)
	}

	got := cache.Users()
	got[0].DisplayName = "mutated"
	if cache.Users()[0].DisplayName != "First" {
		t.Error("Users returned the mutable cache snapshot")
	}
}

func TestUserCacheRefreshFailureAndStickySuppressionPolicy(t *testing.T) {
	providerErr := errors.New("kratos unavailable")
	suppressionErr := errors.New("postgres unavailable")
	providerMode := "ok"
	provider := identityProviderFunc(func(_ context.Context, _ int64, pageToken string) ([]kratosapi.Identity, string, error) {
		switch providerMode {
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

	if err := cache.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got := cache.Users(); len(got) != 1 || got[0].ID != "visible" {
		t.Fatalf("initial users=%+v, want visible only", got)
	}

	providerMode = "later-failed"
	if err := cache.Refresh(t.Context()); !errors.Is(err, providerErr) {
		t.Fatalf("provider error=%v, want Kratos failure", err)
	}
	if got := cache.Users(); len(got) != 1 || got[0].ID != "visible" {
		t.Errorf("provider failure replaced snapshot: %+v", got)
	}

	providerMode = "ok"
	suppressionMode = "failed"
	if err := cache.Refresh(t.Context()); !errors.Is(err, suppressionErr) {
		t.Fatalf("suppression error=%v, want PostgreSQL failure", err)
	}
	if got := cache.Users(); len(got) != 0 {
		t.Errorf("suppression failure retained users: %+v", got)
	}

	suppressionMode = "absent"
	if err := cache.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got := cache.Users(); len(got) != 1 || got[0].ID != "visible" {
		t.Errorf("sticky suppression after successful refresh=%+v, want visible only", got)
	}

	providerMode = "repeated"
	if err := cache.Refresh(t.Context()); err == nil {
		t.Fatal("repeated continuation token was accepted")
	}
	if got := cache.Users(); len(got) != 1 || got[0].ID != "visible" {
		t.Errorf("repeated token replaced complete snapshot: %+v", got)
	}
}
