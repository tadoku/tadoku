package leaderboard

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestDeleteTestTenantRefusesZeroKeyBeforeUsingValkey(t *testing.T) {
	cache := NewCache(nil, time.Second)

	err := cache.DeleteTestTenant(t.Context(), tenant.TestKey{})

	if err == nil {
		t.Fatal("zero test tenant allowed cache deletion")
	}
}

func TestDeleteTestTenantRemovesOnlyOwnedKeysAndIsIdempotent(t *testing.T) {
	client := newLeaderboardTestClient(t)
	suffix := uuid.NewString()[:8]
	key, err := tenant.ParseTestTenant("e2e/cache-" + suffix)
	if err != nil {
		t.Fatal(err)
	}
	otherKey, err := tenant.ParseTestTenant("e2e/other-" + suffix)
	if err != nil {
		t.Fatal(err)
	}
	observed := &deletionCommands{Client: client, t: t}
	cache := NewCache(observed, 5*time.Second)
	prefix, err := cache.cacheKey(tenant.WithKey(t.Context(), key.Key()), "")
	if err != nil {
		t.Fatal(err)
	}
	otherPrefix, err := cache.cacheKey(tenant.WithKey(t.Context(), otherKey.Key()), "")
	if err != nil {
		t.Fatal(err)
	}
	productionPrefix, err := cache.cacheKey(tenant.WithKey(t.Context(), tenant.Production()), "")
	if err != nil {
		t.Fatal(err)
	}

	owned := make([]string, 0, 1102)
	for i := range 1100 {
		owned = append(owned, fmt.Sprintf("%sleaderboard:lifecycle:%d", prefix, i))
	}
	owned = append(owned, prefix+globalKey+":last_updated", prefix+globalKey+":generation")
	untouched := []string{
		productionPrefix + "leaderboard:lifecycle:" + suffix,
		otherPrefix + globalKey,
		otherPrefix + globalKey + ":last_updated",
		otherPrefix + globalKey + ":generation",
		"outside:" + prefix + "marker",
	}
	allKeys := append(append([]string{}, owned...), untouched...)
	var seeded []string
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for start := 0; start < len(seeded); start += 500 {
			end := min(start+500, len(seeded))
			err := client.Do(ctx, client.B().Unlink().Key(seeded[start:end]...).Build()).Error()
			if err != nil {
				t.Error(err)
			}
		}
	})
	for _, item := range allKeys {
		err := client.Do(t.Context(), client.B().Set().Key(item).Value(item).Nx().Build()).Error()
		if err != nil {
			t.Fatalf("seed uniquely owned key %q: %v", item, err)
		}
		seeded = append(seeded, item)
	}

	for attempt := range 2 {
		if err := cache.DeleteTestTenant(t.Context(), key); err != nil {
			t.Fatalf("delete attempt %d: %v", attempt+1, err)
		}

		remaining, err := client.Do(t.Context(), client.B().Exists().Key(owned...).Build()).AsInt64()
		if err != nil {
			t.Fatal(err)
		}
		if remaining != 0 {
			t.Errorf("delete attempt %d left %d tenant-owned keys", attempt+1, remaining)
		}
		for _, item := range untouched {
			value, err := client.Do(t.Context(), client.B().Get().Key(item).Build()).ToString()
			if err != nil || value != item {
				t.Errorf("unrelated key %q = %q, error %v; want unchanged", item, value, err)
			}
		}
	}
	if observed.unlinks < 3 {
		t.Errorf("deleted 1102 keys using %d UNLINK calls; want bounded batches", observed.unlinks)
	}
	t.Logf("removed %d named keys twice; preserved %d unrelated keys", len(owned), len(untouched))
}

type deletionCommands struct {
	valkeygo.Client
	t       *testing.T
	unlinks int
}

func (client *deletionCommands) Do(ctx context.Context, command valkeygo.Completed) valkeygo.ValkeyResult {
	args := command.Commands()
	if args[0] == "UNLINK" {
		client.unlinks++
		if len(args)-1 > 500 {
			client.t.Errorf("UNLINK contains %d keys; maximum is 500", len(args)-1)
		}
	}
	return client.Client.Do(ctx, command)
}
