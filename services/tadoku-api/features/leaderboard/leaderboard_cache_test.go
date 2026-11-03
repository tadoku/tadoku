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
	cache := NewCache(client, time.Second)
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
	if marker != "native:"+after+":1" {
		t.Errorf("marker = %q, want native:%s:1", marker, after)
	}

	if err := cache.invalidate(tenantCtx, key); err != nil {
		t.Fatal(err)
	}
	page, cacheExists, err := cache.fetchPage(tenantCtx, key, 0, 25)
	if err != nil {
		t.Fatal(err)
	}
	if cacheExists || page != nil {
		t.Errorf("cache page after invalidation = %+v, exists = %t; want miss", page, cacheExists)
	}
}

func TestFetchPageServesTheBoardASameGenerationRebuildRepublished(t *testing.T) {
	client := newLeaderboardTestClient(t)
	tenantKey, err := tenant.Parse("e2e/snapshot-" + uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	tenantCtx := tenant.WithKey(t.Context(), tenantKey)
	cache := NewCache(client, time.Second)
	key, err := cache.cacheKey(tenantCtx, globalKey)
	if err != nil {
		t.Fatal(err)
	}
	cleanupLeaderboardKeys(t, client, key)

	first, second, third, fourth := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	before := []score{{userID: first, value: 30}, {userID: second, value: 20}, {userID: fourth, value: 10}}
	after := []score{
		{userID: first, value: 30},
		{userID: second, value: 20},
		{userID: third, value: 15},
		{userID: fourth, value: 10},
	}
	generation, err := cache.generation(tenantCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	published, err := cache.rebuild(tenantCtx, key, before, generation)
	if err != nil || !published {
		t.Fatalf("publish first board: published=%t error=%v", published, err)
	}
	published, err = cache.rebuild(tenantCtx, key, after, generation)
	if err != nil || !published {
		t.Fatalf("republish board: published=%t error=%v", published, err)
	}

	page, exists, err := cache.fetchPage(tenantCtx, key, 2, 1)
	if err != nil || !exists {
		t.Fatalf("fetch page: exists=%t error=%v", exists, err)
	}

	entries := buildEntries(page)
	if len(entries) != 1 || entries[0].Rank != 3 || entries[0].IsTie || entries[0].UserID != third {
		t.Errorf("entries = %+v; want one untied entry at rank 3 from the republished board", entries)
	}
}

func TestRebuildPublishesBoardOf5001Scores(t *testing.T) {
	client := newLeaderboardTestClient(t)
	tenantKey, err := tenant.Parse("e2e/large-" + uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	tenantCtx := tenant.WithKey(t.Context(), tenantKey)
	cache := NewCache(client, 5*time.Second)
	key, err := cache.cacheKey(tenantCtx, globalKey)
	if err != nil {
		t.Fatal(err)
	}
	cleanupLeaderboardKeys(t, client, key)

	generation, err := cache.generation(tenantCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	small := []score{{userID: uuid.New(), value: 1}}
	published, err := cache.rebuild(tenantCtx, key, small, generation)
	if err != nil || !published {
		t.Fatalf("publish small board: published=%t error=%v", published, err)
	}

	large := make([]score, 5001)
	for i := range large {
		large[i] = score{userID: uuid.New(), value: float64(i + 1)}
	}
	published, err = cache.rebuild(tenantCtx, key, large, generation)
	if err != nil || !published {
		t.Errorf("rebuild 5001 scores at the published generation: published=%t error=%v", published, err)
	}

	page, exists, err := cache.fetchPage(tenantCtx, key, 5000, 1)
	if err != nil || !exists {
		t.Fatalf("fetch last page: exists=%t error=%v", exists, err)
	}
	if page.totalCount != 5001 || len(page.scores) != 1 || page.scores[0] != large[0] {
		t.Errorf("last page = %+v; want the lowest of 5001 scores", page)
	}
}

func TestRebuildLimitsMarkerAgeToTwentyEightDays(t *testing.T) {
	client := newLeaderboardTestClient(t)
	tenantKey, err := tenant.Parse("e2e/ttl-" + uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	tenantCtx := tenant.WithKey(t.Context(), tenantKey)
	cache := NewCache(client, time.Second)
	key, err := cache.cacheKey(tenantCtx, globalKey)
	if err != nil {
		t.Fatal(err)
	}
	cleanupLeaderboardKeys(t, client, key)

	if err := cache.invalidate(tenantCtx, key); err != nil {
		t.Fatal(err)
	}
	generation, err := cache.generation(tenantCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	board := []score{{userID: uuid.New(), value: 10}}
	published, err := cache.rebuild(tenantCtx, key, board, generation)
	if err != nil || !published {
		t.Fatalf("publish board: published=%t error=%v", published, err)
	}

	limit := (28 * 24 * time.Hour).Milliseconds()
	markerTTL := pttl(t, client, key+":last_updated")
	if markerTTL <= limit-time.Minute.Milliseconds() || markerTTL > limit {
		t.Errorf("marker PTTL = %d ms; want just under %d ms", markerTTL, limit)
	}
	for _, persistent := range []string{key, key + ":generation"} {
		if ttl := pttl(t, client, persistent); ttl != -1 {
			t.Errorf("%s PTTL = %d; want -1, no expiry", persistent, ttl)
		}
	}
}

func pttl(t *testing.T, client valkeygo.Client, key string) int64 {
	t.Helper()
	ttl, err := client.Do(t.Context(), client.B().Pttl().Key(key).Build()).AsInt64()
	if err != nil {
		t.Fatal(err)
	}
	return ttl
}

func TestFetchPageMissesWhenValkeyLosesTheSortedSet(t *testing.T) {
	client := newLeaderboardTestClient(t)
	cache, tenantCtx, key := newKeyLossFixture(t, client)

	generation, err := cache.generation(tenantCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	board := []score{{userID: uuid.New(), value: 10}}
	published, err := cache.rebuild(tenantCtx, key, board, generation)
	if err != nil || !published {
		t.Fatalf("publish board: published=%t error=%v", published, err)
	}
	loseKey(t, client, key)

	page, exists, err := cache.fetchPage(tenantCtx, key, 0, 25)
	if err != nil || exists {
		t.Errorf("fetch after the set was lost: page=%+v exists=%t error=%v; want a miss", page, exists, err)
	}
}

func TestRebuildIsRefusedAfterValkeyLosesTheGeneration(t *testing.T) {
	client := newLeaderboardTestClient(t)
	cache, tenantCtx, key := newKeyLossFixture(t, client)
	stale := []score{{userID: uuid.New(), value: 10}}

	beforeFirstInvalidation, err := cache.generation(tenantCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.invalidate(tenantCtx, key); err != nil {
		t.Fatal(err)
	}
	loseKey(t, client, key+":generation")
	published, err := cache.rebuild(tenantCtx, key, stale, beforeFirstInvalidation)
	if err != nil || published {
		t.Errorf("rebuild read before the first invalidation: published=%t error=%v; want refused", published, err)
	}

	if err := cache.invalidate(tenantCtx, key); err != nil {
		t.Fatal(err)
	}
	beforeLoss, err := cache.generation(tenantCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	loseKey(t, client, key+":generation")
	if err := cache.invalidate(tenantCtx, key); err != nil {
		t.Fatal(err)
	}
	published, err = cache.rebuild(tenantCtx, key, stale, beforeLoss)
	if err != nil || published {
		t.Errorf("rebuild read before the generation was lost: published=%t error=%v; want refused", published, err)
	}

	current, err := cache.generation(tenantCtx, key)
	if err != nil {
		t.Fatal(err)
	}
	published, err = cache.rebuild(tenantCtx, key, stale, current)
	if err != nil || !published {
		t.Fatalf("publish board: published=%t error=%v", published, err)
	}
	loseKey(t, client, key+":generation")
	page, exists, err := cache.fetchPage(tenantCtx, key, 0, 25)
	if err != nil || exists {
		t.Errorf("fetch after the generation was lost: page=%+v exists=%t error=%v; want a miss", page, exists, err)
	}
}

func newKeyLossFixture(t *testing.T, client valkeygo.Client) (*Cache, context.Context, string) {
	t.Helper()
	tenantKey, err := tenant.Parse("e2e/keyloss-" + uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	tenantCtx := tenant.WithKey(t.Context(), tenantKey)
	cache := NewCache(client, time.Second)
	key, err := cache.cacheKey(tenantCtx, globalKey)
	if err != nil {
		t.Fatal(err)
	}
	cleanupLeaderboardKeys(t, client, key)
	return cache, tenantCtx, key
}

func loseKey(t *testing.T, client valkeygo.Client, key string) {
	t.Helper()
	if err := client.Do(t.Context(), client.B().Del().Key(key).Build()).Error(); err != nil {
		t.Fatal(err)
	}
}

func TestTenantLeaderboardInvalidationKeepsOtherTenantWarm(t *testing.T) {
	client := newLeaderboardTestClient(t)
	cache := NewCache(client, time.Second)
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
	generationA, err := cache.generation(ctxA, cacheA)
	if err != nil {
		t.Fatal(err)
	}
	generationB, err := cache.generation(ctxB, cacheB)
	if err != nil {
		t.Fatal(err)
	}
	published, err := cache.rebuild(ctxA, cacheA, []score{{userID: userA, value: 10}}, generationA)
	if err != nil || !published {
		t.Fatalf("warm tenant A: published=%t error=%v", published, err)
	}
	published, err = cache.rebuild(ctxB, cacheB, []score{{userID: userB, value: 20}}, generationB)
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

func TestGenerationRespectsOperationTimeout(t *testing.T) {
	timeout := 20 * time.Millisecond
	cache := NewCache(&hangingValkey{Client: newLeaderboardTestClient(t)}, timeout)
	started := time.Now()
	_, err := cache.generation(context.Background(), "leaderboard:global")
	elapsed := time.Since(started)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("generation error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed < timeout {
		t.Errorf("generation returned after %v; want at least %v", elapsed, timeout)
	}
	if elapsed > time.Second {
		t.Errorf("generation hung for %v; want operationTimeout to bound it", elapsed)
	}
}

type hangingValkey struct {
	valkeygo.Client
}

func (hangingValkey) Do(ctx context.Context, _ valkeygo.Completed) valkeygo.ValkeyResult {
	<-ctx.Done()
	return valkeygo.NewErrorResult(ctx.Err())
}

func TestLeaderboardWithoutTenantDoesNotWriteCache(t *testing.T) {
	client := newLeaderboardTestClient(t)
	service := NewService(nil, NewCache(client, time.Second))
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
