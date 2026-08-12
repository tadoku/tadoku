package observability

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

const maxAppliedRuleIDs = 10

type scoringShadowLabels struct {
	outcome     domain.ScoringShadowOutcome
	operation   domain.ScoringShadowOperation
	mode        domain.ScoringShadowMode
	activityID  int32
	scoreSource domain.ScoreSource
}

// ScoringShadowMetrics records only values from the approved bounded label
// contract into the service's Prometheus registry.
type ScoringShadowMetrics struct {
	comparisons *prometheus.CounterVec
	enabled     prometheus.Gauge
}

func NewScoringShadowMetrics(registry *prometheus.Registry, engineEnabled bool) *ScoringShadowMetrics {
	comparisons := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tadoku_scoring_shadow_comparisons_total",
			Help: "Legacy-to-engine scoring comparisons.",
		},
		[]string{"outcome", "operation", "mode", "activity_id", "score_source"},
	)
	enabled := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tadoku_scoring_engine_enabled",
		Help: "Whether the scoring engine is authoritative.",
	})
	if engineEnabled {
		enabled.Set(1)
	}
	registry.MustRegister(comparisons, enabled)

	return &ScoringShadowMetrics{comparisons: comparisons, enabled: enabled}
}

func (m *ScoringShadowMetrics) record(observation domain.ScoringShadowObservation) bool {
	labels := scoringShadowLabels{
		outcome:     observation.Outcome,
		operation:   observation.Operation,
		mode:        observation.Mode,
		activityID:  observation.ActivityID,
		scoreSource: observation.ScoreSource,
	}
	if !validScoringShadowLabels(labels) {
		return false
	}

	m.comparisons.WithLabelValues(
		string(labels.outcome),
		string(labels.operation),
		string(labels.mode),
		strconv.FormatInt(int64(labels.activityID), 10),
		string(labels.scoreSource),
	).Inc()
	return true
}

func validScoringShadowLabels(labels scoringShadowLabels) bool {
	if labels.activityID < 1 || labels.activityID > 5 {
		return false
	}
	switch labels.outcome {
	case domain.ScoringShadowOutcomeMatch,
		domain.ScoringShadowOutcomeMismatch,
		domain.ScoringShadowOutcomeUnmatched,
		domain.ScoringShadowOutcomeError:
	default:
		return false
	}
	switch labels.operation {
	case domain.ScoringShadowOperationCreate, domain.ScoringShadowOperationUpdate:
	default:
		return false
	}
	switch labels.mode {
	case domain.ScoringShadowModeShadow, domain.ScoringShadowModeAuthoritative:
	default:
		return false
	}
	switch labels.scoreSource {
	case domain.ScoreSourceAmount, domain.ScoreSourceDurationMinutes:
	default:
		return false
	}
	return true
}

type ScoringShadowObserver struct {
	metrics *ScoringShadowMetrics
	logger  *slog.Logger
}

func NewScoringShadowObserver(metrics *ScoringShadowMetrics, logger *slog.Logger) *ScoringShadowObserver {
	return &ScoringShadowObserver{metrics: metrics, logger: logger}
}

func (o *ScoringShadowObserver) ObserveScoringShadow(ctx context.Context, observation domain.ScoringShadowObservation) {
	if o.metrics == nil || !o.metrics.record(observation) {
		return
	}
	if observation.Outcome == domain.ScoringShadowOutcomeMatch || o.logger == nil {
		return
	}

	attributes := []slog.Attr{
		slog.String("event", "scoring_shadow"),
		slog.String("outcome", string(observation.Outcome)),
		slog.String("operation", string(observation.Operation)),
		slog.String("mode", string(observation.Mode)),
		slog.Int64("activity_id", int64(observation.ActivityID)),
		slog.String("unit_key", observation.UnitKey),
		slog.String("language_code", observation.LanguageCode),
		slog.String("score_source", string(observation.ScoreSource)),
		slog.Float64("legacy_score", float64(observation.LegacyScore)),
	}
	if observation.EngineScore != nil {
		attributes = append(attributes, slog.Float64("engine_score", float64(*observation.EngineScore)))
	}
	if observation.AbsoluteDelta != nil {
		attributes = append(attributes, slog.Float64("absolute_delta", *observation.AbsoluteDelta))
	}
	if observation.RelativeDelta != nil {
		attributes = append(attributes, slog.Float64("relative_delta", *observation.RelativeDelta))
	}
	if observation.RuleSetID != nil {
		attributes = append(attributes, slog.String("rule_set_id", observation.RuleSetID.String()))
	}
	if len(observation.AppliedRuleIDs) > 0 {
		limit := len(observation.AppliedRuleIDs)
		if limit > maxAppliedRuleIDs {
			limit = maxAppliedRuleIDs
		}
		ruleIDs := make([]string, limit)
		for index := 0; index < limit; index++ {
			ruleIDs[index] = observation.AppliedRuleIDs[index].String()
		}
		attributes = append(attributes, slog.Any("applied_rule_ids", ruleIDs))
	}
	if observation.ErrorType != "" {
		attributes = append(attributes, slog.String("error_type", observation.ErrorType))
	}

	level := slog.LevelWarn
	if observation.Outcome == domain.ScoringShadowOutcomeError {
		level = slog.LevelError
	}
	o.logger.LogAttrs(ctx, level, "scoring comparison anomaly", attributes...)
}
