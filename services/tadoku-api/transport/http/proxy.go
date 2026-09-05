package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const correlationHeader = "X-Request-Id"

type Upstreams struct {
	Authz     string
	Content   string
	Immersion string
	Profile   string
}

type route struct {
	name   string
	prefix string
	target string
}

func NewHandler(upstreams Upstreams, transport stdhttp.RoundTripper, requestTimeout time.Duration, registerer prometheus.Registerer, logger *slog.Logger) (stdhttp.Handler, error) {
	if requestTimeout <= 0 {
		return nil, fmt.Errorf("request timeout must be positive")
	}
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	if registerer == nil {
		return nil, fmt.Errorf("metrics registerer is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tadoku_api_proxy_request_duration_seconds",
		Help: "Duration of requests proxied to legacy APIs.",
	}, []string{"route", "upstream", "mode", "status"})
	if err := registerer.Register(duration); err != nil {
		return nil, fmt.Errorf("register proxy metrics: %w", err)
	}

	mux := stdhttp.NewServeMux()
	mux.HandleFunc("GET /livez", func(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = response.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(response stdhttp.ResponseWriter, _ *stdhttp.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"status":"ready","checks":[]}`))
	})

	// The gateway removes its external prefix before forwarding here. Native
	// endpoints can replace these domain proxy routes without that prefix.
	routes := []route{
		{name: "authz", prefix: "/authz/", target: upstreams.Authz},
		{name: "content", prefix: "/content/", target: upstreams.Content},
		{name: "immersion", prefix: "/immersion/", target: upstreams.Immersion},
		{name: "profile", prefix: "/profile/", target: upstreams.Profile},
	}
	for _, current := range routes {
		target, err := parseTarget(current.target)
		if err != nil {
			return nil, fmt.Errorf("%s upstream: %w", current.name, err)
		}
		mux.Handle(current.prefix, observe(current, requestTimeout, newReverseProxy(current, target, transport, logger), duration, logger))
	}

	return mux, nil
}

func parseTarget(raw string) (*url.URL, error) {
	target, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
		return nil, fmt.Errorf("must be an absolute http or https URL")
	}
	if target.User != nil || (target.Path != "" && target.Path != "/") || target.RawQuery != "" || target.Fragment != "" {
		return nil, fmt.Errorf("must not contain credentials, a path, query, or fragment")
	}
	return target, nil
}

func newReverseProxy(current route, target *url.URL, transport stdhttp.RoundTripper, logger *slog.Logger) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(request *httputil.ProxyRequest) {
			request.SetURL(target)
			request.Out.Host = request.In.Host
			request.SetXForwarded()
			request.Out.URL.Path = "/" + strings.TrimPrefix(request.In.URL.Path, current.prefix)
			request.Out.URL.RawPath = "/" + strings.TrimPrefix(request.In.URL.EscapedPath(), current.prefix)
			if request.Out.URL.RawPath == request.Out.URL.Path {
				request.Out.URL.RawPath = ""
			}
		},
		ModifyResponse: func(response *stdhttp.Response) error {
			response.Header.Set(correlationHeader, correlationID(response.Request))
			return nil
		},
		ErrorHandler: func(response stdhttp.ResponseWriter, request *stdhttp.Request, err error) {
			status := stdhttp.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(request.Context().Err(), context.DeadlineExceeded) {
				status = stdhttp.StatusGatewayTimeout
			}
			logger.WarnContext(request.Context(), "proxy request failed",
				"correlation_id", correlationID(request),
				"upstream", current.name,
				"error", err,
			)
			stdhttp.Error(response, stdhttp.StatusText(status), status)
		},
	}
}

func observe(current route, timeout time.Duration, next stdhttp.Handler, duration *prometheus.HistogramVec, logger *slog.Logger) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, request *stdhttp.Request) {
		started := time.Now()
		ctx, cancel := context.WithTimeout(request.Context(), timeout)
		defer cancel()

		request = request.WithContext(withCorrelationID(ctx, request.Header.Get(correlationHeader)))
		request.Header.Set(correlationHeader, correlationID(request))
		response.Header().Set(correlationHeader, correlationID(request))
		recorder := &statusRecorder{ResponseWriter: response}
		next.ServeHTTP(recorder, request)

		status := recorder.status
		if status == 0 {
			status = stdhttp.StatusOK
		}
		elapsed := time.Since(started)
		duration.WithLabelValues(current.prefix, current.name, "proxy", strconv.Itoa(status)).Observe(elapsed.Seconds())
		logger.InfoContext(request.Context(), "request completed",
			"correlation_id", correlationID(request),
			"method", request.Method,
			"route", current.prefix,
			"upstream", current.name,
			"mode", "proxy",
			"status", status,
			"latency", elapsed,
		)
	})
}

type statusRecorder struct {
	stdhttp.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if status >= 100 && status < 200 {
		r.ResponseWriter.WriteHeader(status)
		return
	}
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(stdhttp.StatusOK)
	}
	return r.ResponseWriter.Write(body)
}

func (r *statusRecorder) Unwrap() stdhttp.ResponseWriter { return r.ResponseWriter }

type correlationIDKey struct{}

func withCorrelationID(ctx context.Context, id string) context.Context {
	if id == "" {
		var value [16]byte
		if _, err := rand.Read(value[:]); err == nil {
			id = hex.EncodeToString(value[:])
		} else {
			id = "unavailable"
		}
	}
	return context.WithValue(ctx, correlationIDKey{}, id)
}

func correlationID(request *stdhttp.Request) string {
	id, _ := request.Context().Value(correlationIDKey{}).(string)
	return id
}
