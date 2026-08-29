package featureflags

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type InitializationStatus string

const (
	InitializationStatusReady    InitializationStatus = "ready"
	InitializationStatusFallback InitializationStatus = "fallback"
	InitializationStatusError    InitializationStatus = "error"
)

// Metrics exports only registry-owned flag keys and bounded enum labels. User
// identities and raw provider messages are deliberately absent.
type Metrics struct {
	clock           commondomain.Clock
	evaluations     *prometheus.CounterVec
	duration        *prometheus.HistogramVec
	errors          *prometheus.CounterVec
	initializations *prometheus.CounterVec
	configAge       prometheus.GaugeFunc

	mu                sync.RWMutex
	lastConfigRefresh time.Time
}

func NewMetrics(registry *prometheus.Registry, clock commondomain.Clock) *Metrics {
	m := &Metrics{
		clock: clock,
		evaluations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tadoku_feature_flag_evaluations_total",
			Help: "Boolean feature flag decisions by bounded result metadata.",
		}, []string{"flag_key", "enabled", "reason", "source"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "tadoku_feature_flag_evaluation_duration_seconds",
			Help:    "In-process feature flag evaluation latency.",
			Buckets: prometheus.DefBuckets,
		}, []string{"flag_key", "source"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tadoku_feature_flag_errors_total",
			Help: "Feature flag failures by bounded operation and kind.",
		}, []string{"operation", "kind"}),
		initializations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tadoku_feature_flag_provider_initializations_total",
			Help: "Feature flag provider initialization outcomes.",
		}, []string{"status"}),
	}
	m.configAge = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "tadoku_feature_flag_config_age_seconds",
		Help: "Seconds since the provider last fetched flag configuration successfully.",
	}, m.currentConfigAge)
	registry.MustRegister(m.evaluations, m.duration, m.errors, m.initializations, m.configAge)
	return m
}

func (m *Metrics) ObserveEvaluation(observation Observation) {
	key := observation.Flag.Key()
	if key == "" {
		key = "invalid"
	}
	reason := boundedReason(observation.Reason)
	source := boundedSource(observation.Source)
	m.evaluations.WithLabelValues(
		key,
		boolLabel(observation.Enabled),
		string(reason),
		string(source),
	).Inc()
	m.duration.WithLabelValues(key, string(source)).Observe(observation.Duration.Seconds())
	if observation.Err {
		m.errors.WithLabelValues("evaluation", string(reason)).Inc()
	}
}

func (m *Metrics) ObserveInitialization(status InitializationStatus) {
	switch status {
	case InitializationStatusReady, InitializationStatusFallback, InitializationStatusError:
	default:
		status = InitializationStatusError
	}
	m.initializations.WithLabelValues(string(status)).Inc()
}

func (m *Metrics) ObserveConfigRefresh() {
	m.mu.Lock()
	m.lastConfigRefresh = m.now()
	m.mu.Unlock()
}

func (m *Metrics) ObserveProviderError(kind string) {
	switch kind {
	case "fetch", "initialization":
	default:
		kind = "other"
	}
	m.errors.WithLabelValues("provider", kind).Inc()
}

func (m *Metrics) currentConfigAge() float64 {
	m.mu.RLock()
	lastRefresh := m.lastConfigRefresh
	m.mu.RUnlock()
	if lastRefresh.IsZero() {
		return -1
	}
	age := m.now().Sub(lastRefresh).Seconds()
	if age < 0 {
		return 0
	}
	return age
}

func (m *Metrics) now() time.Time {
	if m.clock == nil {
		return time.Time{}
	}
	return m.clock.Now()
}

func boolLabel(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func boundedReason(reason EvaluationReason) EvaluationReason {
	switch reason {
	case EvaluationReasonMatch,
		EvaluationReasonDefault,
		EvaluationReasonDisabled,
		EvaluationReasonOther,
		EvaluationReasonAnonymous,
		EvaluationReasonInvalidIdentity,
		EvaluationReasonInvalidFlag,
		EvaluationReasonInvalidResponse,
		EvaluationReasonNotFound,
		EvaluationReasonCanceled,
		EvaluationReasonProviderError:
		return reason
	default:
		return EvaluationReasonOther
	}
}

func boundedSource(source EvaluationSource) EvaluationSource {
	switch source {
	case EvaluationSourceProvider, EvaluationSourceStaleCache, EvaluationSourceDefault:
		return source
	default:
		return EvaluationSourceDefault
	}
}
