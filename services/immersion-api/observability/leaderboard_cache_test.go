package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestLeaderboardCacheMetricsExportsBoundedLabelsAndTracksDegradedState(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewLeaderboardCacheMetrics(registry)
	observer := NewLeaderboardCacheObserver(metrics, nil)

	for _, observation := range []domain.LeaderboardCacheObservation{
		{
			Kind:      domain.LeaderboardCacheKindGlobal,
			Operation: domain.LeaderboardCacheOperationFetch,
			Outcome:   domain.LeaderboardCacheOutcomeSuccess,
		},
		{
			Kind:      domain.LeaderboardCacheKindYearly,
			Operation: domain.LeaderboardCacheOperationFetch,
			Outcome:   domain.LeaderboardCacheOutcomeMiss,
		},
		{
			Kind:      domain.LeaderboardCacheKindContest,
			Operation: domain.LeaderboardCacheOperationRebuild,
			Outcome:   domain.LeaderboardCacheOutcomeFailure,
			Err:       errors.New("valkey password must never become a label"),
		},
		{
			Kind:      domain.LeaderboardCacheKindGlobal,
			Operation: domain.LeaderboardCacheOperationUpdate,
			Outcome:   domain.LeaderboardCacheOutcomeFallback,
		},
	} {
		observer.ObserveLeaderboardCache(context.Background(), observation)
	}

	for _, invalid := range []domain.LeaderboardCacheObservation{
		{
			Kind:      domain.LeaderboardCacheKind("user-controlled-kind"),
			Operation: domain.LeaderboardCacheOperationFetch,
			Outcome:   domain.LeaderboardCacheOutcomeFailure,
		},
		{
			Kind:      domain.LeaderboardCacheKindGlobal,
			Operation: domain.LeaderboardCacheOperation("user-controlled-operation"),
			Outcome:   domain.LeaderboardCacheOutcomeFailure,
		},
		{
			Kind:      domain.LeaderboardCacheKindGlobal,
			Operation: domain.LeaderboardCacheOperationFetch,
			Outcome:   domain.LeaderboardCacheOutcome("user-controlled-outcome"),
		},
	} {
		observer.ObserveLeaderboardCache(context.Background(), invalid)
	}

	recorder := httptest.NewRecorder()
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/metrics", nil),
	)
	body := recorder.Body.String()

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, body, `tadoku_leaderboard_cache_operations_total{kind="global",operation="fetch",outcome="success"} 1`)
	assert.Contains(t, body, `tadoku_leaderboard_cache_operations_total{kind="yearly",operation="fetch",outcome="miss"} 1`)
	assert.Contains(t, body, `tadoku_leaderboard_cache_operations_total{kind="contest",operation="rebuild",outcome="failure"} 1`)
	assert.Contains(t, body, `tadoku_leaderboard_cache_operations_total{kind="global",operation="update",outcome="fallback"} 1`)
	assert.Equal(t, 4, strings.Count(body, "tadoku_leaderboard_cache_operations_total{"))
	assert.Contains(t, body, "tadoku_leaderboard_cache_degraded 1")
	assert.NotContains(t, body, "password")
	assert.NotContains(t, body, "user-controlled-kind")
	assert.NotContains(t, body, "user-controlled-operation")
	assert.NotContains(t, body, "user-controlled-outcome")
}

func TestLeaderboardCacheObserverLogsOnlyHealthTransitions(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewLeaderboardCacheMetrics(registry)
	var output bytes.Buffer
	observer := NewLeaderboardCacheObserver(metrics, slog.New(slog.NewJSONHandler(&output, nil)))
	ctx := context.Background()

	observer.ObserveLeaderboardCache(ctx, domain.LeaderboardCacheObservation{
		Kind:      domain.LeaderboardCacheKindGlobal,
		Operation: domain.LeaderboardCacheOperationFetch,
		Outcome:   domain.LeaderboardCacheOutcomeSuccess,
	})
	observer.ObserveLeaderboardCache(ctx, domain.LeaderboardCacheObservation{
		Kind:      domain.LeaderboardCacheKindContest,
		Operation: domain.LeaderboardCacheOperationRebuild,
		Outcome:   domain.LeaderboardCacheOutcomeFailure,
		Err:       errors.New("dial tcp 127.0.0.1:6379: connection refused"),
	})
	observer.ObserveLeaderboardCache(ctx, domain.LeaderboardCacheObservation{
		Kind:      domain.LeaderboardCacheKindYearly,
		Operation: domain.LeaderboardCacheOperationUpdate,
		Outcome:   domain.LeaderboardCacheOutcomeFailure,
		Err:       errors.New("another failure"),
	})
	observer.ObserveLeaderboardCache(ctx, domain.LeaderboardCacheObservation{
		Kind:      domain.LeaderboardCacheKindGlobal,
		Operation: domain.LeaderboardCacheOperationFetch,
		Outcome:   domain.LeaderboardCacheOutcomeMiss,
	})
	observer.ObserveLeaderboardCache(ctx, domain.LeaderboardCacheObservation{
		Kind:      domain.LeaderboardCacheKindGlobal,
		Operation: domain.LeaderboardCacheOperationFetch,
		Outcome:   domain.LeaderboardCacheOutcomeFallback,
	})
	observer.ObserveLeaderboardCache(ctx, domain.LeaderboardCacheObservation{
		Kind:      domain.LeaderboardCacheKindGlobal,
		Operation: domain.LeaderboardCacheOperationFetch,
		Outcome:   domain.LeaderboardCacheOutcomeSuccess,
	})

	lines := nonemptyLines(output.String())
	require.Len(t, lines, 2)

	var degraded map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &degraded))
	assert.Equal(t, "WARN", degraded["level"])
	assert.Equal(t, "leaderboard_cache_health_transition", degraded["event"])
	assert.Equal(t, "degraded", degraded["state"])
	assert.Equal(t, "contest", degraded["kind"])
	assert.Equal(t, "rebuild", degraded["operation"])
	assert.Equal(t, "failure", degraded["outcome"])
	assert.Equal(t, "dial tcp 127.0.0.1:6379: connection refused", degraded["error"])

	var recovered map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &recovered))
	assert.Equal(t, "INFO", recovered["level"])
	assert.Equal(t, "healthy", recovered["state"])
	assert.Equal(t, "success", recovered["outcome"])
	assert.NotContains(t, recovered, "error")

	recorder := httptest.NewRecorder()
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/metrics", nil),
	)
	assert.Contains(t, recorder.Body.String(), "tadoku_leaderboard_cache_degraded 0")
}

func TestLeaderboardCacheObserverIsNilSafe(t *testing.T) {
	observation := domain.LeaderboardCacheObservation{
		Kind:      domain.LeaderboardCacheKindGlobal,
		Operation: domain.LeaderboardCacheOperationFetch,
		Outcome:   domain.LeaderboardCacheOutcomeFailure,
	}

	var observer *LeaderboardCacheObserver
	assert.NotPanics(t, func() {
		observer.ObserveLeaderboardCache(context.Background(), observation)
	})
	assert.NotPanics(t, func() {
		NewLeaderboardCacheObserver(nil, nil).ObserveLeaderboardCache(context.Background(), observation)
	})
}

func TestLeaderboardCacheObserverSerializesConcurrentTransitions(t *testing.T) {
	metrics := NewLeaderboardCacheMetrics(prometheus.NewRegistry())
	var output lockedBuffer
	observer := NewLeaderboardCacheObserver(metrics, slog.New(slog.NewJSONHandler(&output, nil)))

	observeConcurrently(t, 100, func() {
		observer.ObserveLeaderboardCache(context.Background(), domain.LeaderboardCacheObservation{
			Kind:      domain.LeaderboardCacheKindContest,
			Operation: domain.LeaderboardCacheOperationUpdate,
			Outcome:   domain.LeaderboardCacheOutcomeFailure,
			Err:       errors.New("valkey unavailable"),
		})
	})
	observeConcurrently(t, 100, func() {
		observer.ObserveLeaderboardCache(context.Background(), domain.LeaderboardCacheObservation{
			Kind:      domain.LeaderboardCacheKindContest,
			Operation: domain.LeaderboardCacheOperationFetch,
			Outcome:   domain.LeaderboardCacheOutcomeSuccess,
		})
	})

	lines := nonemptyLines(output.String())
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], `"state":"degraded"`)
	assert.Contains(t, lines[1], `"state":"healthy"`)
}

type lockedBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (b *lockedBuffer) Write(value []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(value)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}

func observeConcurrently(t *testing.T, count int, observe func()) {
	t.Helper()
	var waitGroup sync.WaitGroup
	waitGroup.Add(count)
	for range count {
		go func() {
			defer waitGroup.Done()
			observe()
		}()
	}
	waitGroup.Wait()
}

func nonemptyLines(value string) []string {
	lines := make([]string, 0)
	for _, line := range strings.Split(value, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
