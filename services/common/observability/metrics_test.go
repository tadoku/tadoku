package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareUsesNormalizedRouteAndBoundedLabels(t *testing.T) {
	metrics := NewMetrics(nil, "")
	e := echo.New()
	e.Use(metrics.Middleware())
	e.GET("/users/:id", func(ctx echo.Context) error {
		return ctx.NoContent(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/users/private-user-id", nil))

	metricsRecorder := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(metricsRecorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricsRecorder.Body.String()

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Contains(t, body, `http_server_request_duration_seconds_count{error_type="",http_request_method="GET",http_response_status_code="204",http_route="/users/:id"} 1`)
	assert.NotContains(t, body, "private-user-id")
}

func TestMiddlewareRecordsReturnedHTTPErrorStatus(t *testing.T) {
	metrics := NewMetrics(nil, "")
	e := echo.New()
	e.Use(metrics.Middleware())
	e.GET("/unavailable", func(echo.Context) error {
		return echo.NewHTTPError(http.StatusServiceUnavailable)
	})

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/unavailable", nil))

	metricsRecorder := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(metricsRecorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.Contains(t, metricsRecorder.Body.String(), `http_server_request_duration_seconds_count{error_type="server_error",http_request_method="GET",http_response_status_code="503",http_route="/unavailable"} 1`)
}

func TestMiddlewareNeverUsesUnmatchedRawPathAsLabel(t *testing.T) {
	metrics := NewMetrics(nil, "")
	e := echo.New()
	e.Use(metrics.Middleware())

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/missing/private-value", nil))

	metricsRecorder := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(metricsRecorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricsRecorder.Body.String()

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, body, `http_route="unmatched"`)
	assert.NotContains(t, body, "private-value")
}

func TestMetricsExportsRuntimeAndProcessCollectors(t *testing.T) {
	metrics := NewMetrics(nil, "")
	recorder := httptest.NewRecorder()

	metrics.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	metricNames := recorder.Body.String()
	assert.True(t, strings.Contains(metricNames, "go_goroutines ") || strings.Contains(metricNames, "go_sched_goroutines_goroutines "))
	assert.Contains(t, metricNames, "process_cpu_seconds_total")
}
