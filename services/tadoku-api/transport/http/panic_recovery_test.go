package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
)

func TestApplicationPanicRecovery(t *testing.T) {
	var logs bytes.Buffer
	registry := prometheus.NewRegistry()
	router, err := NewHandler(
		app.New(nil, nil, nil, nil, nil, nil, nil, nil, nil),
		func(context.Context) error { return nil },
		time.Second,
		registry,
		slog.New(slog.NewJSONHandler(&logs, nil)),
		func(next stdhttp.Handler) stdhttp.Handler { return next },
		func(next stdhttp.Handler) stdhttp.Handler { return next },
		func(next stdhttp.Handler) stdhttp.Handler { return next },
	)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}
	router.HandleFunc("GET /test/panic", func(stdhttp.ResponseWriter, *stdhttp.Request) {
		panic("test panic")
	})
	router.HandleFunc("GET /test/ok", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	})

	var serverErrors bytes.Buffer
	server := httptest.NewUnstartedServer(router)
	server.Config.ErrorLog = log.New(&serverErrors, "", 0)
	server.Start()
	t.Cleanup(server.Close)

	response, err := server.Client().Get(server.URL + "/test/panic")
	if err != nil {
		t.Fatalf("panicking request: %v", err)
	}
	_, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		t.Errorf("read error = %v, close error = %v", readErr, closeErr)
	}
	if response.StatusCode != stdhttp.StatusInternalServerError {
		t.Errorf("status = %d, want %d", response.StatusCode, stdhttp.StatusInternalServerError)
	}

	healthyResponse, err := server.Client().Get(server.URL + "/test/ok")
	if err != nil {
		t.Fatalf("request after panic: %v", err)
	}
	_ = healthyResponse.Body.Close()
	if healthyResponse.StatusCode != stdhttp.StatusNoContent {
		t.Errorf("status after panic = %d, want %d", healthyResponse.StatusCode, stdhttp.StatusNoContent)
	}

	if got := logs.String(); !strings.Contains(got, `"level":"ERROR"`) ||
		!strings.Contains(got, `"msg":"panic recovered"`) ||
		!strings.Contains(got, `"panic":"test panic"`) ||
		!strings.Contains(got, `"stack":`) {
		t.Errorf("structured logs missing recovered panic and stack: %s", got)
	}
	if got := logs.String(); !strings.Contains(got, `"msg":"request completed"`) ||
		!strings.Contains(got, `"route":"GET /test/panic"`) ||
		!strings.Contains(got, `"mode":"native"`) ||
		!strings.Contains(got, `"status":500`) {
		t.Errorf("structured logs missing observed panic response: %s", got)
	}
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	observed := false
	for _, family := range metricFamilies {
		if family.GetName() != "tadoku_api_proxy_request_duration_seconds" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["route"] == "GET /test/panic" && labels["mode"] == "native" && labels["status"] == "500" {
				observed = metric.GetHistogram().GetSampleCount() == 1
			}
		}
	}
	if !observed {
		t.Error("native panic histogram sample with status 500 not found")
	}
	if got := serverErrors.String(); got != "" {
		t.Errorf("net/http logged recovered panic: %s", got)
	}
}

func TestApplicationPanicRecoveryPreservesWrittenStatus(t *testing.T) {
	var logs bytes.Buffer
	router, err := NewHandler(
		app.New(nil, nil, nil, nil, nil, nil, nil, nil, nil),
		func(context.Context) error { return nil },
		time.Second,
		prometheus.NewRegistry(),
		slog.New(slog.NewJSONHandler(&logs, nil)),
		func(next stdhttp.Handler) stdhttp.Handler { return next },
		func(next stdhttp.Handler) stdhttp.Handler { return next },
		func(next stdhttp.Handler) stdhttp.Handler { return next },
	)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}
	router.HandleFunc("GET /test/partial-panic", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusAccepted)
		_, _ = w.Write([]byte("partial"))
		panic("panic after write")
	})

	var serverErrors bytes.Buffer
	server := httptest.NewUnstartedServer(router)
	server.Config.ErrorLog = log.New(&serverErrors, "", 0)
	server.Start()
	t.Cleanup(server.Close)

	response, err := server.Client().Get(server.URL + "/test/partial-panic")
	if err != nil {
		t.Fatalf("panicking request: %v", err)
	}
	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		t.Errorf("read error = %v, close error = %v", readErr, closeErr)
	}
	if response.StatusCode != stdhttp.StatusAccepted {
		t.Errorf("status = %d, want %d", response.StatusCode, stdhttp.StatusAccepted)
	}
	if string(body) != "partial" {
		t.Errorf("body = %q, want %q", body, "partial")
	}
	if got := serverErrors.String(); got != "" {
		t.Errorf("net/http logged recovered panic or duplicate header: %s", got)
	}
	if got := logs.String(); !strings.Contains(got, `"panic":"panic after write"`) {
		t.Errorf("structured logs missing recovered panic: %s", got)
	}
}

func TestApplicationPanicRecoveryHandlesInvalidStatusPanic(t *testing.T) {
	var logs bytes.Buffer
	router, err := NewHandler(
		app.New(nil, nil, nil, nil, nil, nil, nil, nil, nil),
		func(context.Context) error { return nil },
		time.Second,
		prometheus.NewRegistry(),
		slog.New(slog.NewJSONHandler(&logs, nil)),
		func(next stdhttp.Handler) stdhttp.Handler { return next },
		func(next stdhttp.Handler) stdhttp.Handler { return next },
		func(next stdhttp.Handler) stdhttp.Handler { return next },
	)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}
	router.HandleFunc("GET /test/invalid-status-panic", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(99)
	})

	var serverErrors bytes.Buffer
	server := httptest.NewUnstartedServer(router)
	server.Config.ErrorLog = log.New(&serverErrors, "", 0)
	server.Start()
	t.Cleanup(server.Close)

	response, err := server.Client().Get(server.URL + "/test/invalid-status-panic")
	if err != nil {
		t.Fatalf("panicking request: %v", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != stdhttp.StatusInternalServerError {
		t.Errorf("status = %d, want %d", response.StatusCode, stdhttp.StatusInternalServerError)
	}
	if got := serverErrors.String(); got != "" {
		t.Errorf("net/http logged recovered panic: %s", got)
	}
	if got := logs.String(); !strings.Contains(got, "invalid WriteHeader code 99") {
		t.Errorf("structured logs missing recovered panic: %s", got)
	}
}

func TestApplicationPanicRecoveryTracksFlushedResponse(t *testing.T) {
	var logs bytes.Buffer
	router, err := NewHandler(
		app.New(nil, nil, nil, nil, nil, nil, nil, nil, nil),
		func(context.Context) error { return nil },
		time.Second,
		prometheus.NewRegistry(),
		slog.New(slog.NewJSONHandler(&logs, nil)),
		func(next stdhttp.Handler) stdhttp.Handler { return next },
		func(next stdhttp.Handler) stdhttp.Handler { return next },
		func(next stdhttp.Handler) stdhttp.Handler { return next },
	)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}
	router.HandleFunc("GET /test/flushed-panic", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		if err := stdhttp.NewResponseController(w).Flush(); err != nil {
			panic(err)
		}
		panic("panic after flush")
	})

	var serverErrors bytes.Buffer
	server := httptest.NewUnstartedServer(router)
	server.Config.ErrorLog = log.New(&serverErrors, "", 0)
	server.Start()
	t.Cleanup(server.Close)

	response, err := server.Client().Get(server.URL + "/test/flushed-panic")
	if err != nil {
		t.Fatalf("panicking request: %v", err)
	}
	_, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		t.Errorf("read error = %v, close error = %v", readErr, closeErr)
	}
	if response.StatusCode != stdhttp.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}
	if got := serverErrors.String(); got != "" {
		t.Errorf("net/http logged recovered panic or duplicate header: %s", got)
	}
	if got := logs.String(); !strings.Contains(got, `"panic":"panic after flush"`) {
		t.Errorf("structured logs missing recovered panic: %s", got)
	}
}

func TestApplicationPanicRecoveryTracksSwitchingProtocols(t *testing.T) {
	w := &headerRecordingResponseWriter{header: make(stdhttp.Header)}
	handler := withPanicRecovery(slog.New(slog.NewTextHandler(io.Discard, nil)), stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusSwitchingProtocols)
		panic("panic after protocol switch")
	}))

	handler.ServeHTTP(w, httptest.NewRequest(stdhttp.MethodGet, "/", nil))

	if want := []int{stdhttp.StatusSwitchingProtocols}; !slices.Equal(w.statuses, want) {
		t.Errorf("statuses = %v, want %v", w.statuses, want)
	}
}

func TestApplicationPanicRecoveryIgnoresStatusAfterCommit(t *testing.T) {
	w := &headerRecordingResponseWriter{header: make(stdhttp.Header)}
	handler := withPanicRecovery(slog.New(slog.NewTextHandler(io.Discard, nil)), stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		w.WriteHeader(stdhttp.StatusSwitchingProtocols)
		panic("panic after final status")
	}))

	handler.ServeHTTP(w, httptest.NewRequest(stdhttp.MethodGet, "/", nil))

	if want := []int{stdhttp.StatusOK}; !slices.Equal(w.statuses, want) {
		t.Errorf("statuses = %v, want %v", w.statuses, want)
	}
}

func TestApplicationPanicRecoveryTracksFailedFlush(t *testing.T) {
	flushErr := errors.New("flush failed after commit")
	w := &headerRecordingResponseWriter{
		header:   make(stdhttp.Header),
		flushErr: flushErr,
	}
	handler := withPanicRecovery(slog.New(slog.NewTextHandler(io.Discard, nil)), stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		if err := stdhttp.NewResponseController(w).Flush(); err != nil {
			panic(err)
		}
	}))

	handler.ServeHTTP(w, httptest.NewRequest(stdhttp.MethodGet, "/", nil))

	if want := []int{stdhttp.StatusOK}; !slices.Equal(w.statuses, want) {
		t.Errorf("statuses = %v, want %v", w.statuses, want)
	}
}

func TestApplicationPanicRecoveryPreservesAbortHandler(t *testing.T) {
	var logs bytes.Buffer
	handler := withPanicRecovery(slog.New(slog.NewJSONHandler(&logs, nil)), stdhttp.HandlerFunc(func(stdhttp.ResponseWriter, *stdhttp.Request) {
		panic(stdhttp.ErrAbortHandler)
	}))

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(stdhttp.MethodGet, "/", nil))
	}()

	if recovered != stdhttp.ErrAbortHandler {
		t.Errorf("panic = %v, want http.ErrAbortHandler", recovered)
	}
	if logs.Len() != 0 {
		t.Errorf("abort handler log = %s, want no log", logs.String())
	}
}

type headerRecordingResponseWriter struct {
	header   stdhttp.Header
	statuses []int
	flushErr error
}

func (w *headerRecordingResponseWriter) Header() stdhttp.Header { return w.header }

func (w *headerRecordingResponseWriter) Write(body []byte) (int, error) { return len(body), nil }

func (w *headerRecordingResponseWriter) WriteHeader(status int) {
	w.statuses = append(w.statuses, status)
}

func (w *headerRecordingResponseWriter) FlushError() error {
	if len(w.statuses) == 0 {
		w.statuses = append(w.statuses, stdhttp.StatusOK)
	}
	return w.flushErr
}
