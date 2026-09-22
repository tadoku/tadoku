package observability

import (
	"context"
	"log/slog"
	"math"
	"strconv"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

const maxAppliedRuleIDs = 10

type ScoringComparison struct {
	Operation      string
	Mode           string
	ActivityID     int32
	UnitKey        string
	LanguageCode   string
	ScoreSource    string
	LegacyScore    float32
	EngineScore    *float32
	Matched        bool
	RuleSetID      *uuid.UUID
	AppliedRuleIDs []uuid.UUID
	ErrorType      string
}

type ScoringObserver struct {
	comparisons *prometheus.CounterVec
	enabled     prometheus.Gauge
	logger      *slog.Logger
}

func NewScoringObserver(registry *prometheus.Registry, logger *slog.Logger, enabled bool) *ScoringObserver {
	comparisons := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tadoku_scoring_shadow_comparisons_total",
		Help: "Legacy-to-engine scoring comparisons.",
	}, []string{"outcome", "operation", "mode", "activity_id", "score_source"})
	engineEnabled := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tadoku_scoring_engine_enabled",
		Help: "Whether the scoring engine is authoritative.",
	})
	if enabled {
		engineEnabled.Set(1)
	}
	registry.MustRegister(comparisons, engineEnabled)

	return &ScoringObserver{comparisons: comparisons, enabled: engineEnabled, logger: logger}
}

func (o *ScoringObserver) Observe(ctx context.Context, comparison ScoringComparison) {
	if o == nil || o.comparisons == nil || !validComparison(comparison) {
		return
	}

	outcome := scoringOutcome(comparison)
	o.comparisons.WithLabelValues(
		outcome,
		comparison.Operation,
		comparison.Mode,
		strconv.FormatInt(int64(comparison.ActivityID), 10),
		comparison.ScoreSource,
	).Inc()
	if outcome == "match" || o.logger == nil {
		return
	}

	attributes := []slog.Attr{
		slog.String("event", "scoring_shadow"),
		slog.String("outcome", outcome),
		slog.String("operation", comparison.Operation),
		slog.String("mode", comparison.Mode),
		slog.Int64("activity_id", int64(comparison.ActivityID)),
		slog.String("unit_key", comparison.UnitKey),
		slog.String("language_code", comparison.LanguageCode),
		slog.String("score_source", comparison.ScoreSource),
		slog.Float64("legacy_score", float64(comparison.LegacyScore)),
	}
	if comparison.EngineScore != nil {
		absoluteDelta := math.Abs(float64(comparison.LegacyScore - *comparison.EngineScore))
		scale := math.Max(math.Abs(float64(comparison.LegacyScore)), math.Abs(float64(*comparison.EngineScore)))
		relativeDelta := float64(0)
		if scale > 0 {
			relativeDelta = absoluteDelta / scale
		}
		attributes = append(attributes,
			slog.Float64("engine_score", float64(*comparison.EngineScore)),
			slog.Float64("absolute_delta", absoluteDelta),
			slog.Float64("relative_delta", relativeDelta),
		)
	}
	if comparison.RuleSetID != nil {
		attributes = append(attributes, slog.String("rule_set_id", comparison.RuleSetID.String()))
	}
	if len(comparison.AppliedRuleIDs) > 0 {
		limit := min(len(comparison.AppliedRuleIDs), maxAppliedRuleIDs)
		ruleIDs := make([]string, limit)
		for index := range ruleIDs {
			ruleIDs[index] = comparison.AppliedRuleIDs[index].String()
		}
		attributes = append(attributes, slog.Any("applied_rule_ids", ruleIDs))
	}
	if comparison.ErrorType != "" {
		attributes = append(attributes, slog.String("error_type", comparison.ErrorType))
	}

	level := slog.LevelWarn
	if outcome == "error" {
		level = slog.LevelError
	}
	o.logger.LogAttrs(ctx, level, "scoring comparison anomaly", attributes...)
}

func validComparison(comparison ScoringComparison) bool {
	if comparison.ActivityID < 1 || comparison.ActivityID > 5 {
		return false
	}
	if comparison.Operation != "create" && comparison.Operation != "update" {
		return false
	}
	if comparison.Mode != "shadow" && comparison.Mode != "authoritative" {
		return false
	}
	if comparison.ScoreSource != "amount" && comparison.ScoreSource != "duration_minutes" {
		return false
	}
	switch comparison.ErrorType {
	case "", "scoring_rule_set_not_found", "invalid_scoring_rule_set", "invalid_scoring_input", "evaluation_failed":
		return true
	default:
		return false
	}
}

func scoringOutcome(comparison ScoringComparison) string {
	if comparison.ErrorType != "" {
		return "error"
	}
	if !comparison.Matched {
		return "unmatched"
	}
	if comparison.EngineScore != nil && scoresEqual(comparison.LegacyScore, *comparison.EngineScore) {
		return "match"
	}
	return "mismatch"
}

func scoresEqual(left, right float32) bool {
	difference := math.Abs(float64(left - right))
	scale := math.Max(1, math.Max(math.Abs(float64(left)), math.Abs(float64(right))))
	return difference <= 0.00001*scale
}
