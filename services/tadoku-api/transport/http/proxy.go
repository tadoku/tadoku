package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type Upstreams struct {
	Content   string
	Immersion string
	Profile   string
}

type route struct {
	name   string
	prefix string
	target string
}

// RegisterProxyRoutes attaches temporary legacy routes to the application router.
// Remove this registration when all operations are handled by Tadoku API.
func RegisterProxyRoutes(
	router *Router,
	upstreams Upstreams,
	transport stdhttp.RoundTripper,
	requestTimeout time.Duration,
	logger *slog.Logger,
) error {
	if router == nil || router.rootMux == nil {
		return fmt.Errorf("router is required")
	}
	if requestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}
	if transport == nil {
		return fmt.Errorf("transport is required")
	}
	if logger == nil {
		return fmt.Errorf("logger is required")
	}
	if router.requestDuration == nil {
		return fmt.Errorf("request metrics are required")
	}

	// The gateway removes its external prefix before forwarding here.
	routes := []route{
		{name: "content", prefix: "/content/", target: upstreams.Content},
		{name: "immersion", prefix: "/immersion/", target: upstreams.Immersion},
		{name: "profile", prefix: "/profile/", target: upstreams.Profile},
	}
	// ServeMux GET patterns also match HEAD, and wildcard HEAD exceptions can
	// conflict with more-specific GET patterns. Keep HEAD on the legacy prefixes
	// in a separate proxy-only dispatcher; probes still use the regular root mux.
	headRoutes := stdhttp.NewServeMux()
	headRoutes.Handle("/", router.rootMux)

	for _, current := range routes {
		target, err := parseTarget(current.target)
		if err != nil {
			return fmt.Errorf("%s upstream: %w", current.name, err)
		}

		handler := observe(
			func(*stdhttp.Request) string { return current.prefix },
			current.name,
			"proxy",
			requestTimeout,
			newReverseProxy(current, target, transport, logger),
			router.requestDuration,
			logger,
		)
		router.rootMux.Handle(current.prefix, handler)
		headRoutes.Handle(current.prefix, handler)
	}
	router.rootHandler = stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method == stdhttp.MethodHead {
			headRoutes.ServeHTTP(w, r)
			return
		}
		router.rootMux.ServeHTTP(w, r)
	})

	return nil
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
