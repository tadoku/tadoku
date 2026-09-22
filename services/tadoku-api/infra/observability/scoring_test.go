package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

func TestScoringObserverRecordsBoundedOutcomesAndGauge(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		comparison  ScoringComparison
		wantOutcome string
		wantGauge   float64
	}{
		{
			name:    "matching authoritative score",
			enabled: true,
			comparison: ScoringComparison{
				Operation:   "create",
				Mode:        "authoritative",
				ActivityID:  1,
				ScoreSource: "amount",
				LegacyScore: 10,
				EngineScore: score(10.00009),
				Matched:     true,
			},
			wantOutcome: "match",
			wantGauge:   1,
		},
		{
			name: "mismatching shadow score",
			comparison: ScoringComparison{
				Operation:   "update",
				Mode:        "shadow",
				ActivityID:  2,
				ScoreSource: "duration_minutes",
				LegacyScore: 10,
				EngineScore: score(12),
				Matched:     true,
			},
			wantOutcome: "mismatch",
		},
		{
			name: "unmatched takes precedence over equal scores",
			comparison: ScoringComparison{
				Operation:   "create",
				Mode:        "shadow",
				ActivityID:  3,
				ScoreSource: "amount",
				LegacyScore: 10,
				EngineScore: score(10),
				Matched:     false,
			},
			wantOutcome: "unmatched",
		},
		{
			name: "classified evaluation error",
			comparison: ScoringComparison{
				Operation:   "update",
				Mode:        "authoritative",
				ActivityID:  4,
				ScoreSource: "duration_minutes",
				ErrorType:   "invalid_scoring_input",
			},
			wantOutcome: "error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := prometheus.NewRegistry()
			observer := NewScoringObserver(registry, nil, test.enabled)

			observer.Observe(context.Background(), test.comparison)

			got := metricValue(t, registry, "tadoku_scoring_shadow_comparisons_total", map[string]string{
				"outcome":      test.wantOutcome,
				"operation":    test.comparison.Operation,
				"mode":         test.comparison.Mode,
				"activity_id":  strconv.FormatInt(int64(test.comparison.ActivityID), 10),
				"score_source": test.comparison.ScoreSource,
			})
			if got != 1 {
				t.Errorf("comparison count = %v, want 1", got)
			}
			if got := metricValue(t, registry, "tadoku_scoring_engine_enabled", nil); got != test.wantGauge {
				t.Errorf("enabled gauge = %v, want %v", got, test.wantGauge)
			}
		})
	}
}

func TestScoringObserverRejectsUnboundedMetricLabels(t *testing.T) {
	registry := prometheus.NewRegistry()
	var output bytes.Buffer
	observer := NewScoringObserver(registry, slog.New(slog.NewJSONHandler(&output, nil)), false)

	invalid := []ScoringComparison{
		{Operation: "delete", Mode: "shadow", ActivityID: 1, ScoreSource: "amount", ErrorType: "evaluation_failed"},
		{Operation: "create", Mode: "dry_run", ActivityID: 1, ScoreSource: "amount", ErrorType: "evaluation_failed"},
		{Operation: "create", Mode: "shadow", ActivityID: 0, ScoreSource: "amount", ErrorType: "evaluation_failed"},
		{Operation: "create", Mode: "shadow", ActivityID: 6, ScoreSource: "amount", ErrorType: "evaluation_failed"},
		{Operation: "create", Mode: "shadow", ActivityID: 1, ScoreSource: "user_value", ErrorType: "evaluation_failed"},
		{Operation: "create", Mode: "shadow", ActivityID: 1, ScoreSource: "amount", ErrorType: "database password leaked"},
	}
	for _, comparison := range invalid {
		observer.Observe(context.Background(), comparison)
	}

	if got := metricCount(t, registry, "tadoku_scoring_shadow_comparisons_total"); got != 0 {
		t.Errorf("comparison metric count = %d, want 0", got)
	}
	if output.Len() != 0 {
		t.Errorf("invalid labels produced logs: %s", output.String())
	}
}

func TestScoringObserverLogsSanitizedAnomalyWithFloat32Deltas(t *testing.T) {
	registry := prometheus.NewRegistry()
	var output bytes.Buffer
	observer := NewScoringObserver(registry, slog.New(slog.NewJSONHandler(&output, nil)), false)
	ruleSetID := uuid.New()
	ruleIDs := make([]uuid.UUID, 15)
	for index := range ruleIDs {
		ruleIDs[index] = uuid.New()
	}

	legacyScore := float32(1.1)
	engineScore := float32(12.2)
	observer.Observe(context.Background(), ScoringComparison{
		Operation:      "update",
		Mode:           "shadow",
		ActivityID:     2,
		UnitKey:        "listening_minute",
		LanguageCode:   "jpn",
		ScoreSource:    "duration_minutes",
		LegacyScore:    legacyScore,
		EngineScore:    &engineScore,
		Matched:        true,
		RuleSetID:      &ruleSetID,
		AppliedRuleIDs: ruleIDs,
	})

	event := decodeEvent(t, output.Bytes())
	if event["level"] != "WARN" || event["outcome"] != "mismatch" {
		t.Errorf("anomaly level/outcome = %v/%v, want WARN/mismatch", event["level"], event["outcome"])
	}
	wantAbsolute := math.Abs(float64(legacyScore - engineScore))
	wantRelative := wantAbsolute / math.Max(math.Abs(float64(legacyScore)), math.Abs(float64(engineScore)))
	if event["absolute_delta"] != wantAbsolute || event["relative_delta"] != wantRelative {
		t.Errorf("deltas = %v/%v, want %v/%v", event["absolute_delta"], event["relative_delta"], wantAbsolute, wantRelative)
	}
	if event["rule_set_id"] != ruleSetID.String() {
		t.Errorf("rule_set_id = %v, want %s", event["rule_set_id"], ruleSetID)
	}
	gotRules, ok := event["applied_rule_ids"].([]any)
	if !ok || len(gotRules) != 10 {
		t.Errorf("applied_rule_ids = %#v, want first 10 IDs", event["applied_rule_ids"])
	} else {
		for index, got := range gotRules {
			if got != ruleIDs[index].String() {
				t.Errorf("applied_rule_ids[%d] = %v, want %s", index, got, ruleIDs[index])
			}
		}
	}
	for _, prohibited := range []string{"user_id", "registration_id", "log_id", "tags", "description", "raw_error", "error"} {
		if _, exists := event[prohibited]; exists {
			t.Errorf("log contains prohibited attribute %q", prohibited)
		}
	}
	wantKeys := map[string]bool{
		"time": true, "level": true, "msg": true, "event": true, "outcome": true,
		"operation": true, "mode": true, "activity_id": true, "unit_key": true,
		"language_code": true, "score_source": true, "legacy_score": true,
		"engine_score": true, "absolute_delta": true, "relative_delta": true,
		"rule_set_id": true, "applied_rule_ids": true,
	}
	for key := range event {
		if !wantKeys[key] {
			t.Errorf("log contains unexpected attribute %q", key)
		}
	}
}

func TestScoringObserverLogsErrorsAtErrorAndDoesNotLogMatches(t *testing.T) {
	registry := prometheus.NewRegistry()
	var output bytes.Buffer
	observer := NewScoringObserver(registry, slog.New(slog.NewJSONHandler(&output, nil)), false)
	observer.Observe(context.Background(), ScoringComparison{
		Operation:   "create",
		Mode:        "shadow",
		ActivityID:  1,
		ScoreSource: "amount",
		LegacyScore: 10,
		EngineScore: score(10),
		Matched:     true,
	})
	if output.Len() != 0 {
		t.Fatalf("matching score produced log: %s", output.String())
	}

	observer.Observe(context.Background(), ScoringComparison{
		Operation:   "create",
		Mode:        "shadow",
		ActivityID:  1,
		ScoreSource: "amount",
		ErrorType:   "evaluation_failed",
	})
	event := decodeEvent(t, output.Bytes())
	if event["level"] != "ERROR" || event["error_type"] != "evaluation_failed" {
		t.Errorf("error level/type = %v/%v, want ERROR/evaluation_failed", event["level"], event["error_type"])
	}
	for _, absent := range []string{"engine_score", "absolute_delta", "relative_delta", "rule_set_id", "applied_rule_ids"} {
		if _, exists := event[absent]; exists {
			t.Errorf("error log unexpectedly contains %q", absent)
		}
	}
}

func TestScoringComparisonPrivacyContract(t *testing.T) {
	typeOfComparison := reflect.TypeOf(ScoringComparison{})
	var fields strings.Builder
	for index := 0; index < typeOfComparison.NumField(); index++ {
		fields.WriteString(strings.ToLower(typeOfComparison.Field(index).Name))
		fields.WriteByte(' ')
	}
	for _, prohibited := range []string{"userid", "registrationid", "logid", "tags", "description", "rawerror"} {
		if strings.Contains(fields.String(), prohibited) {
			t.Errorf("comparison contract contains prohibited field %q: %s", prohibited, fields.String())
		}
	}
}

func TestScoringObserverToleranceAndZeroDelta(t *testing.T) {
	tests := []struct {
		name         string
		legacy       float32
		engine       float32
		wantOutcome  string
		wantRelative float64
	}{
		{name: "inside tolerance", legacy: 1, engine: 1.000009, wantOutcome: "match"},
		{name: "outside tolerance", legacy: 1, engine: 1.000011, wantOutcome: "mismatch", wantRelative: math.Abs(float64(float32(1)-float32(1.000011))) / float64(float32(1.000011))},
		{name: "zero denominator", legacy: 0, engine: 0, wantOutcome: "unmatched", wantRelative: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := prometheus.NewRegistry()
			var output bytes.Buffer
			observer := NewScoringObserver(registry, slog.New(slog.NewJSONHandler(&output, nil)), false)
			observer.Observe(context.Background(), ScoringComparison{
				Operation:   "create",
				Mode:        "shadow",
				ActivityID:  1,
				ScoreSource: "amount",
				LegacyScore: test.legacy,
				EngineScore: &test.engine,
				Matched:     test.name != "zero denominator",
			})
			got := metricValue(t, registry, "tadoku_scoring_shadow_comparisons_total", map[string]string{
				"outcome":      test.wantOutcome,
				"operation":    "create",
				"mode":         "shadow",
				"activity_id":  "1",
				"score_source": "amount",
			})
			if got != 1 {
				t.Errorf("comparison count = %v, want 1 for %q", got, test.wantOutcome)
			}
			if test.wantOutcome != "match" {
				event := decodeEvent(t, output.Bytes())
				if event["relative_delta"] != test.wantRelative {
					t.Errorf("relative_delta = %v, want %v", event["relative_delta"], test.wantRelative)
				}
			}
		})
	}
}

func score(value float32) *float32 { return &value }

func decodeEvent(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("decode log event: %v\n%s", err, data)
	}
	return event
}

func metricValue(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) float64 {
	t.Helper()
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.Metric {
			matches := len(metric.Label) == len(labels)
			for _, label := range metric.Label {
				if labels[label.GetName()] != label.GetValue() {
					matches = false
				}
			}
			if matches {
				if metric.Counter != nil {
					return metric.Counter.GetValue()
				}
				return metric.Gauge.GetValue()
			}
		}
	}
	t.Fatalf("metric %q with labels %v not found", name, labels)
	return 0
}

func metricCount(t *testing.T, registry *prometheus.Registry, name string) int {
	t.Helper()
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() == name {
			return len(family.Metric)
		}
	}
	return 0
}
