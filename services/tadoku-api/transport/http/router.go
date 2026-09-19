package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	callbackopenapi "github.com/tadoku/tadoku/services/tadoku-api/generated/openapi/callback"
)

// Router keeps application routes behind shared middleware while allowing this
// package to attach probes and temporary legacy proxies outside it.
type Router struct {
	rootHandler     stdhttp.Handler
	rootMux         *stdhttp.ServeMux
	protectedRoutes *routeRegistrar
	requestDuration *prometheus.HistogramVec
}

type routeRegistrar struct {
	rootMux *stdhttp.ServeMux
	mux     *stdhttp.ServeMux
	handler stdhttp.Handler
}

func (r *routeRegistrar) Handle(pattern string, handler stdhttp.Handler) {
	r.mux.Handle(pattern, handler)
	r.rootMux.Handle(pattern, r.handler)
}

func (r *routeRegistrar) HandleFunc(pattern string, handler func(stdhttp.ResponseWriter, *stdhttp.Request)) {
	r.Handle(pattern, stdhttp.HandlerFunc(handler))
}

func (r *routeRegistrar) ServeHTTP(w stdhttp.ResponseWriter, request *stdhttp.Request) {
	r.mux.ServeHTTP(w, request)
}

type server struct {
	application *app.Application
	logger      *slog.Logger
}

var (
	_ openapi.StrictServerInterface         = (*server)(nil)
	_ callbackopenapi.StrictServerInterface = (*server)(nil)
)

// Handle registers an application route behind the shared middleware.
func (r *Router) Handle(pattern string, handler stdhttp.Handler) {
	r.protectedRoutes.Handle(pattern, handler)
}

// HandleFunc registers an application route behind the shared middleware.
func (r *Router) HandleFunc(pattern string, handler func(stdhttp.ResponseWriter, *stdhttp.Request)) {
	r.Handle(pattern, stdhttp.HandlerFunc(handler))
}

func (r *Router) ServeHTTP(w stdhttp.ResponseWriter, request *stdhttp.Request) {
	r.rootHandler.ServeHTTP(w, request)
}

// NewHandler builds the application router without any legacy upstreams.
func NewHandler(
	application *app.Application,
	ready func(context.Context) error,
	timeout time.Duration,
	registerer prometheus.Registerer,
	logger *slog.Logger,
	authenticate func(stdhttp.Handler) stdhttp.Handler,
	rejectBanned func(stdhttp.Handler) stdhttp.Handler,
	authenticateCallback func(stdhttp.Handler) stdhttp.Handler,
) (*Router, error) {
	if application == nil || ready == nil {
		return nil, fmt.Errorf("application and readiness are required")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("request timeout must be positive")
	}
	if registerer == nil {
		return nil, fmt.Errorf("metrics registerer is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if authenticate == nil {
		return nil, fmt.Errorf("authentication middleware is required")
	}
	if rejectBanned == nil {
		return nil, fmt.Errorf("banned-user middleware is required")
	}
	if authenticateCallback == nil {
		return nil, fmt.Errorf("callback authentication middleware is required")
	}

	requestDuration, err := newRequestDuration(registerer)
	if err != nil {
		return nil, err
	}
	router := &Router{
		rootMux:         stdhttp.NewServeMux(),
		requestDuration: requestDuration,
	}
	router.rootHandler = router.rootMux

	protectedRoutes := &routeRegistrar{
		rootMux: router.rootMux,
		mux:     stdhttp.NewServeMux(),
	}
	applicationHandler := stdhttp.Handler(protectedRoutes)
	applicationHandler = rejectBanned(applicationHandler)
	applicationHandler = authenticate(applicationHandler)
	applicationHandler = withPanicRecovery(logger, applicationHandler)
	protectedRoutes.handler = observe(
		nativeRouteLabel,
		"",
		"native",
		timeout,
		applicationHandler,
		requestDuration,
		logger,
	)
	router.protectedRoutes = protectedRoutes

	callbackRoutes := &routeRegistrar{
		rootMux: router.rootMux,
		mux:     stdhttp.NewServeMux(),
	}
	callbackHandler := authenticateCallback(stdhttp.Handler(callbackRoutes))
	callbackHandler = withPanicRecovery(logger, callbackHandler)
	callbackRoutes.handler = observe(
		nativeRouteLabel,
		"",
		"native",
		timeout,
		callbackHandler,
		requestDuration,
		logger,
	)
	router.rootMux.HandleFunc("GET /livez", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	router.rootMux.Handle("GET /readyz", withRequestTimeout(timeout, readinessHandler(ready)))

	apiServer := &server{
		application: application,
		logger:      logger,
	}
	strictServer := openapi.NewStrictHandlerWithOptions(
		apiServer,
		nil,
		openapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w stdhttp.ResponseWriter, _ *stdhttp.Request, _ error) {
				w.WriteHeader(stdhttp.StatusBadRequest)
			},
			ResponseErrorHandlerFunc: func(w stdhttp.ResponseWriter, request *stdhttp.Request, err error) {
				w.WriteHeader(errorStatus(request.Context(), err))
			},
		},
	)
	openapi.HandlerWithOptions(strictServer, openapi.StdHTTPServerOptions{
		BaseRouter: protectedRoutes,
		Middlewares: []openapi.MiddlewareFunc{
			withJSONCharsetCompatibility,
		},
		ErrorHandlerFunc: func(w stdhttp.ResponseWriter, request *stdhttp.Request, err error) {
			message := err.Error()
			var parameterErr *openapi.InvalidParamFormatError
			if strings.HasPrefix(request.URL.Path, "/immersion/") &&
				errors.As(err, &parameterErr) &&
				strings.HasPrefix(parameterErr.Err.Error(), "error unmarshaling '") {
				message = strings.Replace(message, ": error unmarshaling '", ": error unmarshalling '", 1)
			}
			writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"message": message})
		},
	})
	callbackStrictServer := callbackopenapi.NewStrictHandlerWithOptions(
		apiServer,
		nil,
		callbackopenapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w stdhttp.ResponseWriter, _ *stdhttp.Request, _ error) {
				w.WriteHeader(stdhttp.StatusBadRequest)
			},
			ResponseErrorHandlerFunc: func(w stdhttp.ResponseWriter, request *stdhttp.Request, err error) {
				w.WriteHeader(errorStatus(request.Context(), err))
			},
		},
	)
	callbackopenapi.HandlerWithOptions(callbackStrictServer, callbackopenapi.StdHTTPServerOptions{
		BaseRouter: callbackRoutes,
		Middlewares: []callbackopenapi.MiddlewareFunc{
			withJSONCharsetCompatibility,
		},
		ErrorHandlerFunc: func(w stdhttp.ResponseWriter, _ *stdhttp.Request, err error) {
			writeJSON(w, stdhttp.StatusBadRequest, map[string]string{"message": err.Error()})
		},
	})

	return router, nil
}

func withRequestTimeout(timeout time.Duration, next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func readinessHandler(ready func(context.Context) error) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := ready(r.Context()); err != nil {
			w.WriteHeader(stdhttp.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"not_ready","checks":[{"name":"postgres","status":"failed"}]}`))
			return
		}

		_, _ = w.Write([]byte(`{"status":"ready","checks":[{"name":"postgres","status":"ok"}]}`))
	})
}

func withJSONCharsetCompatibility(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		next.ServeHTTP(&jsonCharsetResponseWriter{ResponseWriter: w}, r)
	})
}

type jsonCharsetResponseWriter struct {
	stdhttp.ResponseWriter
}

func (w *jsonCharsetResponseWriter) WriteHeader(status int) {
	if w.Header().Get("Content-Type") == "application/json" {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *jsonCharsetResponseWriter) Unwrap() stdhttp.ResponseWriter {
	return w.ResponseWriter
}

func writeJSON(w stdhttp.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
