package e2e_test

import (
	"fmt"
	"maps"
	"net/http"
	"testing"
)

func TestImmersionFetchLeaderboardGlobal(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		cache       string
	}{
		{description: []string{"cache", "miss"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"cache", "hit"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"language", "filtered"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"empty", "language"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"tie", "page", "boundary"}, want: http.StatusOK, cache: "hit_tie"},
		{description: []string{"tie", "first", "page"}, want: http.StatusOK, cache: "hit_tie"},
		{description: []string{"cache", "unavailable"}, want: http.StatusOK, cache: "unavailable"},
		{description: []string{"invalid", "activity"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionFetchLeaderboardGlobal", test.want, test.description...)
		runLeaderboardCase(t, name, test.want, test.cache, "leaderboard:global")
	}
}

func TestImmersionFetchLeaderboardForYear(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		cache       string
	}{
		{description: []string{"zero", "score", "cache", "miss"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"cache", "hit"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"activity", "filtered"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"empty", "page"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"empty", "cache", "miss"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"capped", "page", "size"}, want: http.StatusOK, cache: "hit_many"},
		{description: []string{"cache", "unavailable"}, want: http.StatusOK, cache: "unavailable"},
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
	tests := []struct {
		description []string
		want        int
		cache       string
	}{
		{description: []string{"cache", "miss"}, want: http.StatusOK, cache: "miss"},
		{description: []string{"cache", "hit"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"language", "filtered"}, want: http.StatusOK, cache: "hit"},
		{description: []string{"tie", "page", "boundary"}, want: http.StatusOK, cache: "hit_tie"},
		{description: []string{"tie", "first", "page"}, want: http.StatusOK, cache: "hit_tie"},
		{description: []string{"cache", "unavailable"}, want: http.StatusOK, cache: "unavailable"},
		{description: []string{"missing", "organizer"}, want: http.StatusNotFound},
		{description: []string{"missing", "organizer", "filtered"}, want: http.StatusNotFound, cache: "hit"},
		{description: []string{"invalid", "id"}, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		name := APITestName("ImmersionContestFetchLeaderboard", test.want, test.description...)
		cacheKey := "leaderboard:contest:f1111111-1111-4111-8111-111111111111"
		if test.description[0] == "missing" {
			cacheKey = "leaderboard:contest:f2222222-2222-4222-8222-222222222222"
		}
		runLeaderboardCase(t, name, test.want, test.cache, cacheKey)
	}
}

func runLeaderboardCase(t *testing.T, name string, want int, cache, cacheKey string) {
	t.Helper()
	handler := http.Handler(api.handler)
	if cache == "unavailable" {
		handler = unavailableLeaderboardNative
	}
	t.Run(name+"/tadoku-api", func(t *testing.T) {
		dir := "testdata/" + name
		api.reset(t, dir)
		if cache == "hit" || cache == "hit_tie" || cache == "hit_many" {
			seedLeaderboardCache(t, cacheKey, cache)
		}
		atFixtureInstant(func() { checkHTTPGolden(t, handler, dir, want, *updateGoldens) })
		if cache == "miss" {
			verifyRebuiltLeaderboardCache(t, cacheKey)
		}
	})
}

func seedLeaderboardCache(t *testing.T, key, cache string) {
	t.Helper()
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
	if err := client.Do(t.Context(), client.B().Set().Key(key+":last_updated").Value("2026-09-12T12:00:00Z").Build()).Error(); err != nil {
		t.Fatal(err)
	}
}

func verifyRebuiltLeaderboardCache(t *testing.T, key string) {
	t.Helper()
	client := leaderboardValkey.client
	marker, err := client.Do(t.Context(), client.B().Get().Key(key+":last_updated").Build()).ToString()
	if err != nil || marker == "" {
		t.Fatalf("leaderboard rebuild marker = %q, err = %v", marker, err)
	}
	entries, err := client.Do(t.Context(), client.B().Zrange().Key(key).Min("0").Max("-1").Withscores().Build()).AsZScores()
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
	}
	got := make(map[string]float64, len(entries))
	for _, entry := range entries {
		got[entry.Member] = entry.Score
	}
	if !maps.Equal(got, want) {
		t.Fatalf("rebuilt leaderboard entries = %v, want %v", got, want)
	}
}
