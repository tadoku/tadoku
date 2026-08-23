package featureflags

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

func TestMetricsExposeOnlyBoundedLabelsAndConfigAge(t *testing.T) {
	clock := commondomain.NewMockClock(time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC))
	registry := prometheus.NewRegistry()
	metrics := NewMetrics(registry, clock)
	metrics.ObserveInitialization(InitializationStatusFallback)
	metrics.ObserveConfigRefresh()
	clock.SetTime(clock.Now().Add(25 * time.Second))
	metrics.ObserveEvaluation(Observation{
		Flag:     ReleaseLogEntryV2,
		Enabled:  false,
		Reason:   EvaluationReasonProviderError,
		Source:   EvaluationSourceDefault,
		Duration: 4 * time.Millisecond,
		Err:      true,
	})
	metrics.ObserveEvaluation(Observation{
		Flag:    ReleaseLogEntryV2,
		Reason:  EvaluationReason("sensitive@example.com"),
		Source:  EvaluationSource("user-4c47b265"),
		Enabled: true,
	})

	recorder := httptest.NewRecorder()
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := recorder.Body.String()

	assert.Contains(t, body, `tadoku_feature_flag_provider_initializations_total{status="fallback"} 1`)
	assert.Contains(t, body, `tadoku_feature_flag_config_age_seconds 25`)
	assert.Contains(t, body, `tadoku_feature_flag_evaluations_total{enabled="false",flag_key="release.log-entry-v2",reason="provider_error",source="safe_default"} 1`)
	assert.Contains(t, body, `tadoku_feature_flag_errors_total{kind="provider_error",operation="evaluation"} 1`)
	assert.Contains(t, body, `tadoku_feature_flag_evaluations_total{enabled="true",flag_key="release.log-entry-v2",reason="other",source="safe_default"} 1`)
	assert.Equal(t, 2, strings.Count(body, "tadoku_feature_flag_evaluations_total{"))
	assert.NotContains(t, body, "entity")
	assert.NotContains(t, body, "user_id")
	assert.NotContains(t, body, "Sensitive Name")
	assert.NotContains(t, body, "sensitive@example.com")
	assert.NotContains(t, body, "user-4c47b265")
}

func TestMetricsUseSentinelAgeBeforeAnySuccessfulFetch(t *testing.T) {
	registry := prometheus.NewRegistry()
	NewMetrics(registry, commondomain.NewMockClock(time.Time{}))
	recorder := httptest.NewRecorder()
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	assert.Contains(t, recorder.Body.String(), `tadoku_feature_flag_config_age_seconds -1`)
}
