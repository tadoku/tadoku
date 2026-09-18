package http

import (
	"context"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

const correlationHeader = "X-Request-Id"

func newRequestDuration(registerer prometheus.Registerer) (*prometheus.HistogramVec, error) {
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tadoku_api_proxy_request_duration_seconds",
		Help: "Duration of requests handled by Tadoku API.",
	}, []string{"route", "upstream", "mode", "status"})
	if err := registerer.Register(duration); err != nil {
		return nil, fmt.Errorf("register request metrics: %w", err)
	}
	return duration, nil
}

func observe(
	routeLabel func(*stdhttp.Request) string,
	upstream string,
	mode string,
	timeout time.Duration,
	next stdhttp.Handler,
	duration *prometheus.HistogramVec,
	logger *slog.Logger,
) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, request *stdhttp.Request) {
		started := time.Now()
		ctx, cancel := context.WithTimeout(request.Context(), timeout)
		defer cancel()

		request = request.WithContext(withCorrelationID(ctx, request.Header.Get(correlationHeader)))
		request.Header.Set(correlationHeader, correlationID(request))
		if mode == "proxy" {
			response.Header().Set(correlationHeader, correlationID(request))
		}
		recorder := &statusRecorder{ResponseWriter: response}
		next.ServeHTTP(recorder, request)

		status := recorder.status
		if status == 0 {
			status = stdhttp.StatusOK
		}
		elapsed := time.Since(started)
		route := routeLabel(request)
		duration.WithLabelValues(route, upstream, mode, strconv.Itoa(status)).Observe(elapsed.Seconds())
		logger.InfoContext(request.Context(), "request completed",
			"correlation_id", correlationID(request),
			"method", request.Method,
			"route", route,
			"upstream", upstream,
			"mode", mode,
			"status", status,
			"latency", elapsed,
		)
	})
}

func nativeRouteLabel(request *stdhttp.Request) string {
	if request.Pattern == "" {
		return "unmatched"
	}
	return request.Pattern
}

type statusRecorder struct {
	stdhttp.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	if status >= 100 && status < 200 && status != stdhttp.StatusSwitchingProtocols {
		r.ResponseWriter.WriteHeader(status)
		return
	}
	r.ResponseWriter.WriteHeader(status)
	r.status = status
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(stdhttp.StatusOK)
	}
	return r.ResponseWriter.Write(body)
}

func (r *statusRecorder) Unwrap() stdhttp.ResponseWriter { return r.ResponseWriter }

func (r *statusRecorder) FlushError() error {
	if r.status == 0 {
		r.status = stdhttp.StatusOK
	}
	return stdhttp.NewResponseController(r.ResponseWriter).Flush()
}

type correlationIDKey struct{}

func withCorrelationID(ctx context.Context, id string) context.Context {
	if id == "" {
		id = uuid.Must(uuid.NewV7()).String()
	}
	return context.WithValue(ctx, correlationIDKey{}, id)
}

func correlationID(request *stdhttp.Request) string {
	id, _ := request.Context().Value(correlationIDKey{}).(string)
	return id
}
