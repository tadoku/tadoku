package e2e_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestImmersionFetchLeaderboardGlobal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description []string
		want        int
		cache       string
	}{
		{description: []string{"cache", "miss"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"cache", "miss", "past", "end"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"cache", "hit"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"language", "filtered"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"activity", "filtered"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"empty", "language"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"tie", "page", "boundary"}, want: http.StatusOK, cache: "hit_tie"},
		{description: []string{"tie", "first", "page"}, want: http.StatusOK, cache: "hit_tie"},
		{description: []string{"tie", "order", "across", "paths"}, want: http.StatusOK, cache: "miss_then_hit"},
		{description: []string{"invalid", "activity"}, want: http.StatusBadRequest},
		{description: []string{"negative", "page"}, want: http.StatusBadRequest},
		{description: []string{"negative", "page", "size"}, want: http.StatusBadRequest},
		{description: []string{"offset", "overflow"}, want: http.StatusOK},
		{description: []string{"cache", "hit", "negative", "page"}, want: http.StatusBadRequest, cache: "hit"},
		{description: []string{"cache", "hit", "offset", "overflow"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"cache", "hit", "maximum", "page"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"overflowing", "activity"}, want: http.StatusBadRequest},
		{description: []string{"negative", "overflowing", "activity"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionFetchLeaderboardGlobal", test.want, test.description...)
		runLeaderboardCase(t, name, test.want, test.cache, "leaderboard:global")
	}
}

func TestImmersionFetchLeaderboardForYear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description []string
		want        int
		cache       string
	}{
		{description: []string{"zero", "score", "cache", "miss"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"zero", "score", "cache", "hit"}, want: http.StatusOK, cache: "miss_then_hit"},
		{description: []string{"tie", "order", "across", "paths"}, want: http.StatusOK, cache: "miss_then_hit"},
		{description: []string{"cache", "miss", "past", "end"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"cache", "hit"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"activity", "filtered"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"empty", "page"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"empty", "cache", "miss"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"capped", "page", "size"}, want: http.StatusOK, cache: "hit_many"},
		{description: []string{"negative", "page"}, want: http.StatusBadRequest},
		{description: []string{"offset", "overflow"}, want: http.StatusOK},
		{description: []string{"cache", "hit", "offset", "overflow"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"overflowing", "activity"}, want: http.StatusBadRequest},
		{description: []string{"negative", "overflowing", "activity"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionFetchLeaderboardForYear", test.want, test.description...)
		cacheKey := "leaderboard:yearly:2026"
		if test.description[0] == "empty" && test.description[1] == "cache" {
			cacheKey = "leaderboard:yearly:2024"
		}
		runLeaderboardCase(t, name, test.want, test.cache, cacheKey)
	}
}

func TestImmersionContestFetchLeaderboard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description []string
		want        int
		cache       string
	}{
		{description: []string{"cache", "miss"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"cache", "miss", "past", "end"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"cache", "hit"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"language", "filtered"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"tie", "page", "boundary"}, want: http.StatusOK, cache: "hit_tie"},
		{description: []string{"tie", "first", "page"}, want: http.StatusOK, cache: "hit_tie"},
		{description: []string{"tie", "order", "across", "paths"}, want: http.StatusOK, cache: "miss_then_hit"},
		{description: []string{"deleted", "participant"}, want: http.StatusOK, cache: "miss_then_hit"},
		{description: []string{"deleted", "participant", "activity", "filtered"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"missing", "organizer"}, want: http.StatusNotFound},
		{description: []string{"missing", "organizer", "filtered"}, want: http.StatusNotFound, cache: "hit"},
		{description: []string{"missing", "contest", "cache", "miss"}, want: http.StatusNotFound, cache: "miss_then_miss"},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"negative", "page"}, want: http.StatusBadRequest},
		{description: []string{"offset", "overflow"}, want: http.StatusOK},
		{description: []string{"cache", "hit", "offset", "overflow"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"overflowing", "activity"}, want: http.StatusBadRequest},
		{description: []string{"negative", "overflowing", "activity"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionContestFetchLeaderboard", test.want, test.description...)
		cacheKey := "leaderboard:contest:f1111111-1111-4111-8111-111111111111"
		if test.description[0] == "deleted" {
			cacheKey = "leaderboard:contest:f3333333-3333-4333-8333-333333333333"
		}
		if test.description[0] == "missing" {
			cacheKey = "leaderboard:contest:f2222222-2222-4222-8222-222222222222"
			if len(test.description) > 1 && test.description[1] == "contest" {
				cacheKey = "leaderboard:contest:f9999999-9999-4999-8999-999999999999"
			}
		}
		runLeaderboardCase(t, name, test.want, test.cache, cacheKey)
	}
}

func TestImmersionContestFetchLeaderboardEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description []string
		want        int
		cache       string
		cacheKey    string
	}{
		{description: []string{"rank", "below", "fifty"}, want: http.StatusOK, cache: "miss_then_hit"},
		{description: []string{"tie", "below", "fifty"}, want: http.StatusOK, cache: "miss_then_hit"},
		{description: []string{"registered", "without", "logs"}, want: http.StatusOK, cache: "miss_then_hit"},
		{
			description: []string{"cache", "hit", "tie"},
			want:        http.StatusOK,
			cache:       "hit_tie",
			cacheKey:    "leaderboard:contest:f1111111-1111-4111-8111-111111111111",
		},
		{description: []string{"not", "registered"}, want: http.StatusNoContent, cache: "miss_then_hit"},
		{description: []string{"withdrawn", "registration"}, want: http.StatusNoContent, cache: "miss_then_hit"},
		{
			description: []string{"missing", "contest"},
			want:        http.StatusNotFound,
			cache:       "miss_then_miss",
			cacheKey:    "leaderboard:contest:f9999999-9999-4999-8999-999999999999",
		},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "user", "id"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionContestFetchLeaderboardEntry", test.want, test.description...)
		cacheKey := test.cacheKey
		if cacheKey == "" {
			cacheKey = "leaderboard:contest:f4444444-4444-4444-8444-444444444444"
		}
		runLeaderboardCase(t, name, test.want, test.cache, cacheKey)
	}
}

// runLeaderboardCase runs the case in its own fixtureScenario, in parallel with other scenario cases, with cache
// describing the primary tenant's cacheKey before and after the request.
func runLeaderboardCase(t *testing.T, name string, want int, cache, cacheKey string) {
	t.Helper()
	t.Run(name+"/tadoku-api", func(t *testing.T) {
		t.Parallel()

		dir := filepath.Join("testdata", name)
		impl := implementation{name: "tadoku-api", handler: api.handler}
		s := newFixtureScenario(t)
		tenantKey := s.tenant(t, primaryRole).String()
		if cache == "hit" || cache == "hit_tie" || cache == "hit_many" {
			seedLeaderboardCache(t, tenantKey, cacheKey, cache)
		}
		s.runCase(t, dir, want, fixtureInstant, impl)
		if cache == "miss" || cache == "miss_then_hit" {
			verifyRebuiltLeaderboardCache(t, tenantKey, cacheKey)
		}
		if cache == "miss_then_hit" || cache == "miss_then_miss" {
			s.checkCase(t, dir, want, impl)
		}
	})
}

// seedLeaderboardCache warms cacheKey with a board built from the tenant's current version of it, so the next read
// serves the seeded scores.
func seedLeaderboardCache(t *testing.T, tenantKey, cacheKey, cache string) {
	t.Helper()
	board := strings.TrimPrefix(cacheKey, "leaderboard:")
	key := "tenant:" + tenantKey + ":" + cacheKey
	client := leaderboardValkey.client
	builder := client.B().Zadd().Key(key).ScoreMember().ScoreMember(77, "11111111-1111-4111-8111-111111111111")
	if cache == "hit_many" {
		builder = client.B().Zadd().Key(key).ScoreMember().ScoreMember(101, "80000000-0000-4000-8000-000000000001")
	}
	if cache == "hit_tie" {
		builder = builder.ScoreMember(20, "33333333-3333-4333-8333-333333333333").ScoreMember(20, "22222222-2222-4222-8222-222222222222").ScoreMember(0, "66666666-6666-4666-8666-666666666666")
	} else if cache == "hit_many" {
		for i := 2; i <= 101; i++ {
			builder = builder.ScoreMember(float64(102-i), fmt.Sprintf("80000000-0000-4000-8000-%012d", i))
		}
	}
	if err := client.Do(t.Context(), builder.Build()).Error(); err != nil {
		t.Fatal(err)
	}
	size, err := client.Do(t.Context(), client.B().Zcard().Key(key).Build()).AsInt64()
	if err != nil {
		t.Fatal(err)
	}
	version := strconv.FormatInt(leaderboardVersion(t, tenantKey, board), 10)
	marker := client.B().Set().
		Key(key + ":last_updated").
		Value("version:" + version + ":" + strconv.FormatInt(size, 10)).
		Build()
	if err := client.Do(t.Context(), marker).Error(); err != nil {
		t.Fatal(err)
	}
}

func cachedLeaderboardScore(t *testing.T, tenantKey, key string) float64 {
	t.Helper()
	key = "tenant:" + tenantKey + ":" + key
	entries, err := leaderboardValkey.client.Do(t.Context(), leaderboardValkey.client.B().Zrange().Key(key).Min("0").Max("-1").Withscores().Build()).AsZScores()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Member == "11111111-1111-4111-8111-111111111111" {
			return entry.Score
		}
	}
	t.Fatalf("user score absent from cache %s", key)
	return 0
}

func verifyRebuiltLeaderboardCache(t *testing.T, tenantKey, key string) {
	t.Helper()
	cacheKey := "tenant:" + tenantKey + ":" + key
	client := leaderboardValkey.client
	marker, err := client.Do(t.Context(), client.B().Get().Key(cacheKey+":last_updated").Build()).ToString()
	if err != nil || marker == "" {
		t.Fatalf("leaderboard rebuild marker = %q, err = %v", marker, err)
	}
	entries, err := client.Do(
		t.Context(),
		client.B().Zrange().Key(cacheKey).Min("0").Max("-1").Withscores().Build(),
	).AsZScores()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]float64{}
	switch key {
	case "leaderboard:global":
		want = map[string]float64{
			"11111111-1111-4111-8111-111111111111": 30,
			"22222222-2222-4222-8222-222222222222": 20,
			"33333333-3333-4333-8333-333333333333": 20,
			"55555555-5555-4555-8555-555555555555": 10,
			"66666666-6666-4666-8666-666666666666": 5,
		}
	case "leaderboard:yearly:2026":
		want = map[string]float64{
			"11111111-1111-4111-8111-111111111111": 30,
			"22222222-2222-4222-8222-222222222222": 20,
			"33333333-3333-4333-8333-333333333333": 20,
			"55555555-5555-4555-8555-555555555555": 10,
		}
	case "leaderboard:contest:f1111111-1111-4111-8111-111111111111":
		want = map[string]float64{
			"11111111-1111-4111-8111-111111111111": 30,
			"22222222-2222-4222-8222-222222222222": 20,
			"33333333-3333-4333-8333-333333333333": 20,
			"66666666-6666-4666-8666-666666666666": 0,
		}
	case "leaderboard:contest:f4444444-4444-4444-8444-444444444444":
		want["80000000-0000-4000-8000-000000000061"] = 0
		for i := 1; i <= 60; i++ {
			want[fmt.Sprintf("80000000-0000-4000-8000-%012d", i)] = float64(61 - i)
		}
		want["80000000-0000-4000-8000-000000000059"] = 3
	case "leaderboard:contest:f3333333-3333-4333-8333-333333333333":
		want = map[string]float64{
			"11111111-1111-4111-8111-111111111111": 10,
			"22222222-2222-4222-8222-222222222222": 25,
			"77777777-7777-4777-8777-777777777777": 40,
			"88888888-8888-4888-8888-888888888888": 0,
		}
	}
	got := make(map[string]float64, len(entries))
	for _, entry := range entries {
		got[entry.Member] = entry.Score
	}
	if !maps.Equal(got, want) {
		t.Fatalf("rebuilt leaderboard entries = %v, want %v", got, want)
	}
}

func leaderboardVersion(t *testing.T, tenantKey, board string) int64 {
	t.Helper()
	var version int64
	err := pgx.BeginTxFunc(t.Context(), api.db.AppPool, pgx.TxOptions{AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		if _, err := tx.Exec(t.Context(), "select set_config('tadoku.tenant', $1, true)", tenantKey); err != nil {
			return err
		}
		return tx.QueryRow(
			t.Context(),
			"select coalesce((select version from leaderboard_versions where board = $1), 0)",
			board,
		).Scan(&version)
	})
	if err != nil {
		t.Fatal(err)
	}
	return version
}

func TestImmersionLeaderboardsServeAWriteOnTheNextRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cacheKey string
		write    string
	}{
		{
			name:     "ImmersionFetchLeaderboardGlobal/200_cache_miss",
			cacheKey: "leaderboard:global",
			write:    "update logs set computed_score = 110 where id = 'a1111111-1111-4111-8111-000000000001'",
		},
		{
			name:     "ImmersionFetchLeaderboardForYear/200_zero_score_cache_miss",
			cacheKey: "leaderboard:yearly:2026",
			write:    "update logs set computed_score = 110 where id = 'a1111111-1111-4111-8111-000000000001'",
		},
		{
			name:     "ImmersionContestFetchLeaderboard/200_cache_miss",
			cacheKey: "leaderboard:contest:f1111111-1111-4111-8111-111111111111",
			write:    "update contest_logs set computed_score = 110 where log_id = 'a1111111-1111-4111-8111-000000000001'",
		},
	}
	for _, test := range tests {
		t.Run(test.cacheKey, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join("testdata", test.name)
			impl := implementation{name: "tadoku-api", handler: api.handler}
			s := newFixtureScenario(t)
			tenantKey := s.tenant(t, primaryRole).String()
			s.runCase(t, dir, http.StatusOK, fixtureInstant, impl)
			verifyRebuiltLeaderboardCache(t, tenantKey, test.cacheKey)

			err := pgx.BeginFunc(t.Context(), api.db.AppPool, func(tx pgx.Tx) error {
				if _, err := tx.Exec(t.Context(), "select set_config('tadoku.tenant', $1, true)", tenantKey); err != nil {
					return err
				}
				_, err := tx.Exec(t.Context(), test.write)
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			request := readHTTPRequest(t, dir)
			defer request.Body.Close()
			recorder := s.serveFixture(t, impl.handler, request, "")

			var board struct {
				Entries []struct {
					UserID string  `json:"user_id"`
					Score  float64 `json:"score"`
				} `json:"entries"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &board); err != nil {
				t.Fatalf("decode %d response %q: %v", recorder.Code, recorder.Body.String(), err)
			}
			if len(board.Entries) == 0 || board.Entries[0].UserID != "11111111-1111-4111-8111-111111111111" ||
				board.Entries[0].Score != 130 {
				t.Errorf("entries after the write = %+v; want user 11111111-… first with 130", board.Entries)
			}
			if got := cachedLeaderboardScore(t, tenantKey, test.cacheKey); got != 130 {
				t.Errorf("cached score after the next read = %v; want 130", got)
			}
		})
	}
}
