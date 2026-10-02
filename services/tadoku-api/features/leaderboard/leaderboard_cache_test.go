package leaderboard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestRebuildDoesNotPublishSnapshotAfterInvalidation(t *testing.T) {
	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	option.SelectDB = 13
	option.ForceSingleClient = true
	option.DisableRetry = true
	client, err := valkeygo.NewClient(option)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)

	tenantKey, err := tenant.Parse("e2e/fence-" + uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	tenantCtx := tenant.WithKey(t.Context(), tenantKey)
	cache := NewCache(client, time.Second, "")
	key, err := cache.cacheKey(tenantCtx, globalKey)
	if err != nil {
		t.Fatal(err)
	}
	lease := key + ":lease"
	if err := client.Do(tenantCtx, client.B().Set().Key(lease).Value("test").Nx().Build()).Error(); err != nil {
		if errors.Is(err, valkeygo.Nil) {
			t.Fatal("test key is leased")
		}
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := client.Do(ctx, client.B().Del().Key(key, key+":last_updated", key+":generation", lease).Build()).Error(); err != nil {
			t.Error(err)
		}
	})

	before, err := cache.generation(tenantCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.invalidate(tenantCtx, key); err != nil {
		t.Fatal(err)
	}
	stale := []score{{userID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), value: 10}}
	published, err := cache.rebuild(tenantCtx, key, stale, before)
	if err != nil {
		t.Fatal(err)
	}
	if published {
		t.Error("stale snapshot was published after invalidation")
	}
	exists, err := client.Do(tenantCtx, client.B().Exists().Key(key+":last_updated").Build()).AsInt64()
	if err != nil {
		t.Fatal(err)
	}
	if exists != 0 {
		t.Errorf("stale marker exists after rejected rebuild")
	}

	after, err := cache.generation(tenantCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	fresh := []score{{userID: stale[0].userID, value: 20}}
	published, err = cache.rebuild(tenantCtx, key, fresh, after)
	if err != nil {
		t.Fatal(err)
	}
	if !published {
		t.Error("fresh snapshot was rejected")
	}
	marker, err := client.Do(tenantCtx, client.B().Get().Key(key+":last_updated").Build()).ToString()
	if err != nil {
		t.Fatal(err)
	}
	if marker != "native:"+after {
		t.Errorf("marker = %q, want native:%s", marker, after)
	}

	racing := NewCache(&invalidateBeforeZcard{Client: client, key: key, t: t}, time.Second, "")
	page, cacheExists, err := racing.fetchPage(tenantCtx, key, 0, 25)
	if err != nil {
		t.Fatal(err)
	}
	if cacheExists || page != nil {
		t.Errorf("cache page after invalidation = %+v, exists = %t; want miss", page, cacheExists)
	}
}

type invalidateBeforeZcard struct {
	valkeygo.Client
	key   string
	t     *testing.T
	calls int
}

func (client *invalidateBeforeZcard) Do(ctx context.Context, command valkeygo.Completed) valkeygo.ValkeyResult {
	client.calls++
	if client.calls == 3 {
		if err := invalidateScript.Exec(ctx, client.Client, []string{client.key, client.key + ":last_updated", client.key + ":generation"}, nil).Error(); err != nil {
			client.t.Fatal(err)
		}
	}
	return client.Client.Do(ctx, command)
}

func TestTenantLeaderboardInvalidationKeepsOtherTenantWarm(t *testing.T) {
	client := newLeaderboardTestClient(t)
	cache := NewCache(client, time.Second, "")
	service := NewService(nil, cache)
	contestID := uuid.New()
	keyA, err := tenant.Parse("e2e/cache-a-" + uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	keyB, err := tenant.Parse("e2e/cache-b-" + uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	ctxA := tenant.WithKey(t.Context(), keyA)
	ctxB := tenant.WithKey(t.Context(), keyB)

	cacheA, err := cache.cacheKey(ctxA, contestPrefix+contestID.String())
	if err != nil {
		t.Fatal(err)
	}
	cacheB, err := cache.cacheKey(ctxB, contestPrefix+contestID.String())
	if err != nil {
		t.Fatal(err)
	}
	cleanupLeaderboardKeys(t, client, cacheA, cacheB)
	if cacheA != "tenant:"+keyA.String()+":"+contestPrefix+contestID.String() ||
		cacheB != "tenant:"+keyB.String()+":"+contestPrefix+contestID.String() {
		t.Fatalf("tenant leaderboard keys: A=%q B=%q", cacheA, cacheB)
	}

	userA, userB := uuid.New(), uuid.New()
	published, err := cache.rebuild(ctxA, cacheA, []score{{userID: userA, value: 10}}, "0")
	if err != nil || !published {
		t.Fatalf("warm tenant A: published=%t error=%v", published, err)
	}
	published, err = cache.rebuild(ctxB, cacheB, []score{{userID: userB, value: 20}}, "0")
	if err != nil || !published {
		t.Fatalf("warm tenant B: published=%t error=%v", published, err)
	}

	if err := service.InvalidateContest(ctxA, contestID); err != nil {
		t.Fatal(err)
	}

	page, exists, err := cache.fetchPage(ctxB, cacheB, 0, 25)
	if err != nil || !exists || page == nil {
		t.Fatalf("tenant B warm cache after A invalidation: page=%+v exists=%t error=%v", page, exists, err)
	}
	if len(page.scores) != 1 || page.scores[0].userID != userB || page.scores[0].value != 20 {
		t.Errorf("tenant B warm scores=%+v; want its own score", page.scores)
	}
	page, exists, err = cache.fetchPage(ctxA, cacheA, 0, 25)
	if err != nil || exists || page != nil {
		t.Errorf("tenant A cache after invalidation: page=%+v exists=%t error=%v", page, exists, err)
	}
}

func TestLeaderboardWithoutTenantDoesNotWriteCache(t *testing.T) {
	client := newLeaderboardTestClient(t)
	service := NewService(nil, NewCache(client, time.Second, ""))
	contestID := uuid.New()
	key := contestPrefix + contestID.String()
	cleanupLeaderboardKeys(t, client, key)

	err := service.InvalidateContest(t.Context(), contestID)
	if err == nil {
		t.Error("missing tenant allowed leaderboard invalidation")
	}

	count, err := client.Do(
		t.Context(),
		client.B().Exists().Key(key, key+":last_updated", key+":generation").Build(),
	).AsInt64()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("missing tenant wrote %d cache keys", count)
	}
}

func newLeaderboardTestClient(t *testing.T) valkeygo.Client {
	t.Helper()
	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	option.SelectDB = 13
	option.ForceSingleClient = true
	option.DisableRetry = true

	client, err := valkeygo.NewClient(option)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	return client
}

func cleanupLeaderboardKeys(t *testing.T, client valkeygo.Client, keys ...string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		for _, key := range keys {
			err := client.Do(ctx, client.B().Del().Key(key, key+":last_updated", key+":generation").Build()).Error()
			if err != nil {
				t.Error(err)
			}
		}
	})
}
