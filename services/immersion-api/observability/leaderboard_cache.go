package observability

import (
	"context"
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

// LeaderboardCacheMetrics contains process-local leaderboard cache health
// metrics. Labels are limited to the bounded values in the domain contract.
type LeaderboardCacheMetrics struct {
	operations *prometheus.CounterVec
	degraded   prometheus.Gauge
}

func NewLeaderboardCacheMetrics(registry *prometheus.Registry) *LeaderboardCacheMetrics {
	operations := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tadoku_leaderboard_cache_operations_total",
			Help: "Leaderboard cache operations by leaderboard kind, operation, and outcome.",
		},
		[]string{"kind", "operation", "outcome"},
	)
	degraded := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tadoku_leaderboard_cache_degraded",
		Help: "Whether this immersion-api process has observed the leaderboard cache as degraded.",
	})
	degraded.Set(0)
	registry.MustRegister(operations, degraded)

	return &LeaderboardCacheMetrics{
		operations: operations,
		degraded:   degraded,
	}
}

func (m *LeaderboardCacheMetrics) record(observation domain.LeaderboardCacheObservation) {
	if m == nil || m.operations == nil {
		return
	}
	m.operations.WithLabelValues(
		string(observation.Kind),
		string(observation.Operation),
		string(observation.Outcome),
	).Inc()
}

func (m *LeaderboardCacheMetrics) setDegraded(degraded bool) {
	if m == nil || m.degraded == nil {
		return
	}
	if degraded {
		m.degraded.Set(1)
		return
	}
	m.degraded.Set(0)
}

// LeaderboardCacheObserver exports cache metrics and emits one structured log
// for each process-local health transition.
type LeaderboardCacheObserver struct {
	metrics *LeaderboardCacheMetrics
	logger  *slog.Logger

	mu       sync.Mutex
	degraded bool
}

func NewLeaderboardCacheObserver(metrics *LeaderboardCacheMetrics, logger *slog.Logger) *LeaderboardCacheObserver {
	return &LeaderboardCacheObserver{metrics: metrics, logger: logger}
}

func (o *LeaderboardCacheObserver) ObserveLeaderboardCache(ctx context.Context, observation domain.LeaderboardCacheObservation) {
	if o == nil || !validLeaderboardCacheObservation(observation) {
		return
	}

	o.metrics.record(observation)

	degraded, establishesHealth := leaderboardCacheHealth(observation.Outcome)
	if !establishesHealth {
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	if o.degraded == degraded {
		return
	}
	o.degraded = degraded
	o.metrics.setDegraded(degraded)
	o.logTransition(ctx, observation, degraded)
}

func (o *LeaderboardCacheObserver) logTransition(
	ctx context.Context,
	observation domain.LeaderboardCacheObservation,
	degraded bool,
) {
	if o.logger == nil {
		return
	}

	state := "healthy"
	message := "leaderboard cache recovered"
	level := slog.LevelInfo
	if degraded {
		state = "degraded"
		message = "leaderboard cache degraded"
		level = slog.LevelWarn
	}

	attributes := []slog.Attr{
		slog.String("event", "leaderboard_cache_health_transition"),
		slog.String("state", state),
		slog.String("kind", string(observation.Kind)),
		slog.String("operation", string(observation.Operation)),
		slog.String("outcome", string(observation.Outcome)),
	}
	if observation.Err != nil {
		attributes = append(attributes, slog.Any("error", observation.Err))
	}
	o.logger.LogAttrs(ctx, level, message, attributes...)
}

func validLeaderboardCacheObservation(observation domain.LeaderboardCacheObservation) bool {
	switch observation.Kind {
	case domain.LeaderboardCacheKindGlobal,
		domain.LeaderboardCacheKindYearly,
		domain.LeaderboardCacheKindContest:
	default:
		return false
	}

	switch observation.Operation {
	case domain.LeaderboardCacheOperationFetch,
		domain.LeaderboardCacheOperationRebuild,
		domain.LeaderboardCacheOperationUpdate:
	default:
		return false
	}

	switch observation.Outcome {
	case domain.LeaderboardCacheOutcomeSuccess,
		domain.LeaderboardCacheOutcomeMiss,
		domain.LeaderboardCacheOutcomeFailure,
		domain.LeaderboardCacheOutcomeFallback:
		return true
	default:
		return false
	}
}

func leaderboardCacheHealth(outcome domain.LeaderboardCacheOutcome) (degraded, establishesHealth bool) {
	switch outcome {
	case domain.LeaderboardCacheOutcomeSuccess:
		return false, true
	case domain.LeaderboardCacheOutcomeFailure:
		return true, true
	default:
		return false, false
	}
}
