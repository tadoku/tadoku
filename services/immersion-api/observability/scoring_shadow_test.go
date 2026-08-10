package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestScoringShadowMetricsExportsBoundedLabelsAndExactCounts(t *testing.T) {
	metrics := NewScoringShadowMetrics(true)
	observer := NewScoringShadowObserver(metrics, nil)
	observation := domain.ScoringShadowObservation{
		Outcome:      domain.ScoringShadowOutcomeMatch,
		Operation:    domain.ScoringShadowOperationCreate,
		Mode:         domain.ScoringShadowModeAuthoritative,
		ActivityID:   1,
		ScoreSource:  domain.ScoreSourceAmount,
		LegacyScore:  10,
		LanguageCode: "jpn",
	}

	observer.ObserveScoringShadow(context.Background(), observation)
	observer.ObserveScoringShadow(context.Background(), observation)
	invalid := observation
	invalid.ActivityID = 9001
	invalid.ScoreSource = domain.ScoreSource("user-controlled-source")
	observer.ObserveScoringShadow(context.Background(), invalid)

	recorder := httptest.NewRecorder()
	metrics.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := recorder.Body.String()

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Contains(t, body, `tadoku_scoring_shadow_comparisons_total{outcome="match",operation="create",mode="authoritative",activity_id="1",score_source="amount"} 2`)
	assert.Equal(t, 1, strings.Count(body, "tadoku_scoring_shadow_comparisons_total{"))
	assert.Contains(t, body, "tadoku_scoring_engine_enabled 1")
	assert.NotContains(t, body, "language_code")
	assert.NotContains(t, body, "unit_key")
	assert.NotContains(t, body, "user-controlled-source")
}

func TestScoringShadowMetricsRejectsNonGETRequests(t *testing.T) {
	metrics := NewScoringShadowMetrics(false)
	recorder := httptest.NewRecorder()

	metrics.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/metrics", nil))

	assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	assert.Equal(t, http.MethodGet, recorder.Header().Get("Allow"))
}

func TestScoringShadowObserverWritesSanitizedBoundedJSONAnomalies(t *testing.T) {
	metrics := NewScoringShadowMetrics(false)
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	observer := NewScoringShadowObserver(metrics, logger)
	ruleSetID := uuid.New()
	engineScore := float32(12)
	absoluteDelta := float64(2)
	relativeDelta := float64(1) / 6
	ruleIDs := make([]uuid.UUID, maxAppliedRuleIDs+5)
	for index := range ruleIDs {
		ruleIDs[index] = uuid.New()
	}

	observer.ObserveScoringShadow(context.Background(), domain.ScoringShadowObservation{
		Outcome:        domain.ScoringShadowOutcomeMismatch,
		Operation:      domain.ScoringShadowOperationUpdate,
		Mode:           domain.ScoringShadowModeShadow,
		ActivityID:     2,
		UnitKey:        domain.UnitKeyListeningMinute,
		LanguageCode:   "jpn",
		ScoreSource:    domain.ScoreSourceDurationMinutes,
		LegacyScore:    10,
		EngineScore:    &engineScore,
		AbsoluteDelta:  &absoluteDelta,
		RelativeDelta:  &relativeDelta,
		RuleSetID:      &ruleSetID,
		AppliedRuleIDs: ruleIDs,
	})

	var event map[string]any
	require.NoError(t, json.Unmarshal(output.Bytes(), &event))
	assert.Equal(t, "scoring_shadow", event["event"])
	assert.Equal(t, "mismatch", event["outcome"])
	assert.Equal(t, "update", event["operation"])
	assert.Equal(t, "shadow", event["mode"])
	assert.Equal(t, float64(2), event["activity_id"])
	assert.Equal(t, domain.UnitKeyListeningMinute, event["unit_key"])
	assert.Equal(t, "jpn", event["language_code"])
	assert.Equal(t, "duration_minutes", event["score_source"])
	assert.Equal(t, float64(10), event["legacy_score"])
	assert.Equal(t, float64(12), event["engine_score"])
	assert.Equal(t, ruleSetID.String(), event["rule_set_id"])
	assert.Len(t, event["applied_rule_ids"], maxAppliedRuleIDs)

	for _, prohibited := range []string{
		"user_id",
		"registration_id",
		"log_id",
		"tags",
		"description",
		"uri",
		"headers",
		"request_body",
		"error",
	} {
		assert.NotContains(t, event, prohibited)
	}
}

func TestScoringShadowObserverDoesNotLogMatchesOrRawErrors(t *testing.T) {
	metrics := NewScoringShadowMetrics(false)
	var output bytes.Buffer
	observer := NewScoringShadowObserver(metrics, slog.New(slog.NewJSONHandler(&output, nil)))
	base := domain.ScoringShadowObservation{
		Operation:    domain.ScoringShadowOperationCreate,
		Mode:         domain.ScoringShadowModeShadow,
		ActivityID:   1,
		ScoreSource:  domain.ScoreSourceAmount,
		LanguageCode: "jpn",
	}
	match := base
	match.Outcome = domain.ScoringShadowOutcomeMatch
	observer.ObserveScoringShadow(context.Background(), match)
	assert.Empty(t, output.String())

	failure := base
	failure.Outcome = domain.ScoringShadowOutcomeError
	failure.ErrorType = "evaluation_failed"
	observer.ObserveScoringShadow(context.Background(), failure)
	assert.NotContains(t, output.String(), "password")
	assert.NotContains(t, output.String(), "database")
	assert.Contains(t, output.String(), `"error_type":"evaluation_failed"`)
}

func TestScoringShadowObservationPrivacyContractHasNoProhibitedFields(t *testing.T) {
	typeOfObservation := reflect.TypeOf(domain.ScoringShadowObservation{})
	fields := make([]string, 0, typeOfObservation.NumField())
	for index := 0; index < typeOfObservation.NumField(); index++ {
		fields = append(fields, strings.ToLower(typeOfObservation.Field(index).Name))
	}
	joined := strings.Join(fields, " ")

	for _, prohibited := range []string{"userid", "registrationid", "logid", "tags", "description", "uri", "header", "requestbody", "rawerror"} {
		assert.NotContains(t, joined, prohibited)
	}
}

func TestMetricsServerStartsAndShutsDownGracefully(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	handler := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-releaseRequest
		response.WriteHeader(http.StatusNoContent)
	})
	server := NewMetricsServer("127.0.0.1:0", handler)
	require.NoError(t, server.Start())

	requestDone := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + server.Addr())
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			err = response.Body.Close()
		}
		requestDone <- err
	}()
	<-requestStarted

	shutdownDone := make(chan error, 1)
	shutdownContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() {
		shutdownDone <- server.Shutdown(shutdownContext)
	}()

	select {
	case err := <-shutdownDone:
		require.Failf(t, "shutdown returned before active request completed", "error: %v", err)
	case <-time.After(25 * time.Millisecond):
	}
	close(releaseRequest)
	require.NoError(t, <-requestDone)
	require.NoError(t, <-shutdownDone)
	_, err := http.Get("http://" + server.Addr())
	assert.Error(t, err)
}
