package observability

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"sync"

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

// ScoringShadowMetrics is both a bounded in-memory metric registry and a
// Prometheus text exposition handler. Only values from the approved label
// contract are accepted.
type ScoringShadowMetrics struct {
	mu          sync.RWMutex
	comparisons map[scoringShadowLabels]uint64
	enabled     bool
}

func NewScoringShadowMetrics(engineEnabled bool) *ScoringShadowMetrics {
	return &ScoringShadowMetrics{
		comparisons: make(map[scoringShadowLabels]uint64),
		enabled:     engineEnabled,
	}
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

	m.mu.Lock()
	m.comparisons[labels]++
	m.mu.Unlock()
	return true
}

func (m *ScoringShadowMetrics) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		response.Header().Set("Allow", http.MethodGet)
		http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	m.mu.RLock()
	labels := make([]scoringShadowLabels, 0, len(m.comparisons))
	for label := range m.comparisons {
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool {
		return metricLabelSortKey(labels[i]) < metricLabelSortKey(labels[j])
	})

	_, _ = fmt.Fprintln(response, "# HELP tadoku_scoring_shadow_comparisons_total Legacy-to-engine scoring comparisons.")
	_, _ = fmt.Fprintln(response, "# TYPE tadoku_scoring_shadow_comparisons_total counter")
	for _, label := range labels {
		_, _ = fmt.Fprintf(
			response,
			"tadoku_scoring_shadow_comparisons_total{outcome=%q,operation=%q,mode=%q,activity_id=%q,score_source=%q} %d\n",
			label.outcome,
			label.operation,
			label.mode,
			strconv.FormatInt(int64(label.activityID), 10),
			label.scoreSource,
			m.comparisons[label],
		)
	}
	_, _ = fmt.Fprintln(response, "# HELP tadoku_scoring_engine_enabled Whether the scoring engine is authoritative.")
	_, _ = fmt.Fprintln(response, "# TYPE tadoku_scoring_engine_enabled gauge")
	if m.enabled {
		_, _ = fmt.Fprintln(response, "tadoku_scoring_engine_enabled 1")
	} else {
		_, _ = fmt.Fprintln(response, "tadoku_scoring_engine_enabled 0")
	}
	m.mu.RUnlock()
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

func metricLabelSortKey(labels scoringShadowLabels) string {
	return string(labels.outcome) + "\x00" +
		string(labels.operation) + "\x00" +
		string(labels.mode) + "\x00" +
		strconv.FormatInt(int64(labels.activityID), 10) + "\x00" +
		string(labels.scoreSource)
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
