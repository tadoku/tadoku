package observability

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const unmatchedRoute = "unmatched"

var httpDurationBuckets = []float64{
	0.005,
	0.01,
	0.025,
	0.05,
	0.1,
	0.25,
	0.5,
	1,
	2.5,
	5,
	10,
}

type Metrics struct {
	registry            *prometheus.Registry
	httpRequestDuration *prometheus.HistogramVec
}

func NewMetrics(db *sql.DB, databaseName string) *Metrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	if db != nil {
		registry.MustRegister(collectors.NewDBStatsCollector(db, databaseName))
	}

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_server_request_duration_seconds",
			Help:    "Duration of inbound HTTP server requests.",
			Buckets: httpDurationBuckets,
		},
		[]string{
			"http_request_method",
			"http_route",
			"http_response_status_code",
			"error_type",
		},
	)
	registry.MustRegister(httpRequestDuration)

	return &Metrics{
		registry:            registry,
		httpRequestDuration: httpRequestDuration,
	}
}

func (m *Metrics) Registerer() prometheus.Registerer {
	return m.registry
}

func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Metrics) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			started := time.Now()
			err := next(ctx)

			status := ctx.Response().Status
			if status == 0 {
				status = http.StatusOK
			}
			if err != nil && status < http.StatusBadRequest {
				status = statusCode(err)
			}

			route := ctx.Path()
			if route == "" || (status == http.StatusNotFound && route == ctx.Request().URL.Path) {
				route = unmatchedRoute
			}

			errorType := ""
			if status >= http.StatusInternalServerError {
				errorType = "server_error"
			}

			m.httpRequestDuration.WithLabelValues(
				ctx.Request().Method,
				route,
				strconv.Itoa(status),
				errorType,
			).Observe(time.Since(started).Seconds())

			return err
		}
	}
}

func statusCode(err error) int {
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		return httpError.Code
	}
	return http.StatusInternalServerError
}
