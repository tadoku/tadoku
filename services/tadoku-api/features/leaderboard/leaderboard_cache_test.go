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

func TestFetchPageServesOnlyTheVersionTheBoardWasBuiltFrom(t *testing.T) {
	client := newLeaderboardTestClient(t)
	cache, tenantCtx, key := newCacheFixture(t, client)

	board := []score{{userID: uuid.New(), value: 10}}
	published, err := cache.rebuild(tenantCtx, key, board, 3)
	if err != nil || !published {
		t.Fatalf("publish board at version 3: published=%t error=%v", published, err)
	}

	for _, version := range []int64{2, 3, 4} {
		page, exists, err := cache.fetchPage(tenantCtx, key, version, 0, 25)
		if err != nil {
			t.Fatal(err)
		}
		if exists != (version == 3) {
			t.Errorf("fetch at version %d: page=%+v exists=%t; want a hit only at version 3", version, page, exists)
		}
	}
}

func TestRebuildKeepsABoardBuiltFromANewerVersion(t *testing.T) {
	tests := []struct {
		name         string
		newer, older int64
	}{
		{name: "small", newer: 5, older: 4},
		{name: "beyond float precision", newer: 1<<53 + 1, older: 1 << 53},
		{name: "more digits", newer: 10, older: 9},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := newLeaderboardTestClient(t)
			cache, tenantCtx, key := newCacheFixture(t, client)
			user := uuid.New()

			published, err := cache.rebuild(tenantCtx, key, []score{{userID: user, value: 20}}, test.newer)
			if err != nil || !published {
				t.Fatalf("publish board at version %d: published=%t error=%v", test.newer, published, err)
			}
			published, err = cache.rebuild(tenantCtx, key, []score{{userID: user, value: 10}}, test.older)
			if err != nil || published {
				t.Errorf("rebuild at version %d over %d: published=%t error=%v; want refused",
					test.older,
					test.newer,
					published,
					err,
				)
			}

			page, exists, err := cache.fetchPage(tenantCtx, key, test.newer, 0, 25)
			if err != nil || !exists {
				t.Fatalf("fetch at version %d: exists=%t error=%v", test.newer, exists, err)
			}
			if len(page.scores) != 1 || page.scores[0].value != 20 {
				t.Errorf("scores at version %d = %+v; want the newer board", test.newer, page.scores)
			}
		})
	}
}

func TestFetchPageServesTheBoardASameVersionRebuildRepublished(t *testing.T) {
	client := newLeaderboardTestClient(t)
	cache, tenantCtx, key := newCacheFixture(t, client)

	first, second, third, fourth := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	before := []score{{userID: first, value: 30}, {userID: second, value: 20}, {userID: fourth, value: 10}}
	after := []score{
		{userID: first, value: 30},
		{userID: second, value: 20},
		{userID: third, value: 15},
		{userID: fourth, value: 10},
	}
	published, err := cache.rebuild(tenantCtx, key, before, 1)
	if err != nil || !published {
		t.Fatalf("publish first board: published=%t error=%v", published, err)
	}
	published, err = cache.rebuild(tenantCtx, key, after, 1)
	if err != nil || !published {
		t.Fatalf("republish board: published=%t error=%v", published, err)
	}

	page, exists, err := cache.fetchPage(tenantCtx, key, 1, 2, 1)
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
	cache, tenantCtx, key := newCacheFixture(t, client)
	cache.operationTimeout = 5 * time.Second

	small := []score{{userID: uuid.New(), value: 1}}
	published, err := cache.rebuild(tenantCtx, key, small, 1)
	if err != nil || !published {
		t.Fatalf("publish small board: published=%t error=%v", published, err)
	}

	large := make([]score, 5001)
	for i := range large {
		large[i] = score{userID: uuid.New(), value: float64(i + 1)}
	}
	published, err = cache.rebuild(tenantCtx, key, large, 1)
	if err != nil || !published {
		t.Errorf("rebuild 5001 scores at the published version: published=%t error=%v", published, err)
	}

	page, exists, err := cache.fetchPage(tenantCtx, key, 1, 5000, 1)
	if err != nil || !exists {
		t.Fatalf("fetch last page: exists=%t error=%v", exists, err)
	}
	if page.totalCount != 5001 || len(page.scores) != 1 || page.scores[0] != large[0] {
		t.Errorf("last page = %+v; want the lowest of 5001 scores", page)
	}
}

func TestRebuildSetsNoExpiry(t *testing.T) {
	client := newLeaderboardTestClient(t)
	cache, tenantCtx, key := newCacheFixture(t, client)

	board := []score{{userID: uuid.New(), value: 10}}
	published, err := cache.rebuild(tenantCtx, key, board, 1)
	if err != nil || !published {
		t.Fatalf("publish board: published=%t error=%v", published, err)
	}

	for _, item := range []string{key, key + ":last_updated"} {
		if ttl := pttl(t, client, item); ttl != -1 {
			t.Errorf("%s PTTL = %d; want -1, no expiry", item, ttl)
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

func TestFetchPageMissesWhenValkeyLosesAKeyUntilARebuild(t *testing.T) {
	for _, suffix := range []string{"", ":last_updated"} {
		t.Run("lost"+suffix, func(t *testing.T) {
			client := newLeaderboardTestClient(t)
			cache, tenantCtx, key := newCacheFixture(t, client)
			board := []score{{userID: uuid.New(), value: 10}}
			published, err := cache.rebuild(tenantCtx, key, board, 1)
			if err != nil || !published {
				t.Fatalf("publish board: published=%t error=%v", published, err)
			}

			loseKey(t, client, key+suffix)
			page, exists, err := cache.fetchPage(tenantCtx, key, 1, 0, 25)
			if err != nil || exists {
				t.Errorf("fetch after the loss: page=%+v exists=%t error=%v; want a miss", page, exists, err)
			}
			published, err = cache.rebuild(tenantCtx, key, board, 1)
			if err != nil || !published {
				t.Fatalf("rebuild after the loss: published=%t error=%v", published, err)
			}
			if _, exists, err := cache.fetchPage(tenantCtx, key, 1, 0, 25); err != nil || !exists {
				t.Errorf("fetch after the rebuild: exists=%t error=%v; want a hit", exists, err)
			}
		})
	}
}

func TestFetchPageMissesABoardThePreviousReleasePublished(t *testing.T) {
	client := newLeaderboardTestClient(t)
	cache, tenantCtx, key := newCacheFixture(t, client)
	user := uuid.New().String()
	previousRelease := []valkeygo.Completed{
		client.B().Zadd().Key(key).ScoreMember().ScoreMember(10, user).Build(),
		client.B().Set().Key(key + ":generation").Value("7").Build(),
		client.B().Set().Key(key + ":last_updated").Value("native:7:1").Build(),
	}
	for _, command := range previousRelease {
		if err := client.Do(t.Context(), command).Error(); err != nil {
			t.Fatal(err)
		}
	}

	for _, version := range []int64{0, 1, 7} {
		if page, exists, err := cache.fetchPage(tenantCtx, key, version, 0, 25); err != nil || exists {
			t.Errorf("fetch at version %d: page=%+v exists=%t error=%v; want a miss", version, page, exists, err)
		}
	}
	published, err := cache.rebuild(tenantCtx, key, []score{{userID: uuid.New(), value: 20}}, 1)
	if err != nil || !published {
		t.Errorf("rebuild over the previous release's board: published=%t error=%v", published, err)
	}
}

func newCacheFixture(t *testing.T, client valkeygo.Client) (*Cache, context.Context, string) {
	t.Helper()
	tenantKey, err := tenant.Parse("e2e/cache-" + uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	tenantCtx := tenant.WithKey(t.Context(), tenantKey)
	cache := NewCache(client, time.Second)
	key, err := cache.boardKey(tenantCtx, globalBoard)
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

	cacheA, err := cache.boardKey(ctxA, contestBoard(contestID))
	if err != nil {
		t.Fatal(err)
	}
	cacheB, err := cache.boardKey(ctxB, contestBoard(contestID))
	if err != nil {
		t.Fatal(err)
	}
	cleanupLeaderboardKeys(t, client, cacheA, cacheB)
	if cacheA != "tenant:"+keyA.String()+":leaderboard:contest:"+contestID.String() ||
		cacheB != "tenant:"+keyB.String()+":leaderboard:contest:"+contestID.String() {
		t.Fatalf("tenant leaderboard keys: A=%q B=%q", cacheA, cacheB)
	}

	userA, userB := uuid.New(), uuid.New()
	published, err := cache.rebuild(ctxA, cacheA, []score{{userID: userA, value: 10}}, 1)
	if err != nil || !published {
		t.Fatalf("warm tenant A: published=%t error=%v", published, err)
	}
	published, err = cache.rebuild(ctxB, cacheB, []score{{userID: userB, value: 20}}, 1)
	if err != nil || !published {
		t.Fatalf("warm tenant B: published=%t error=%v", published, err)
	}

	if err := service.InvalidateContest(ctxA, contestID); err != nil {
		t.Fatal(err)
	}

	page, exists, err := cache.fetchPage(ctxB, cacheB, 1, 0, 25)
	if err != nil || !exists || page == nil {
		t.Fatalf("tenant B warm cache after A invalidation: page=%+v exists=%t error=%v", page, exists, err)
	}
	if len(page.scores) != 1 || page.scores[0].userID != userB || page.scores[0].value != 20 {
		t.Errorf("tenant B warm scores=%+v; want its own score", page.scores)
	}
	page, exists, err = cache.fetchPage(ctxA, cacheA, 1, 0, 25)
	if err != nil || exists || page != nil {
		t.Errorf("tenant A cache after invalidation: page=%+v exists=%t error=%v", page, exists, err)
	}
}

func TestFetchPageRespectsOperationTimeout(t *testing.T) {
	timeout := 20 * time.Millisecond
	cache := NewCache(&hangingValkey{Client: newLeaderboardTestClient(t)}, timeout)
	started := time.Now()
	_, _, err := cache.fetchPage(context.Background(), "leaderboard:global", 1, 0, 25)
	elapsed := time.Since(started)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("fetchPage error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed < timeout {
		t.Errorf("fetchPage returned after %v; want at least %v", elapsed, timeout)
	}
	if elapsed > time.Second {
		t.Errorf("fetchPage hung for %v; want operationTimeout to bound it", elapsed)
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
	key := "leaderboard:contest:" + contestID.String()
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
