package http

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type receivedRequest struct {
	method        string
	path          string
	rawQuery      string
	body          string
	authorization string
	correlationID string
	host          string
}

func TestHandlerProxiesEachLegacyPrefix(t *testing.T) {
	received := make(map[string]chan receivedRequest)
	servers := make(map[string]*httptest.Server)
	for _, name := range []string{"authz", "content", "immersion", "profile"} {
		name := name
		received[name] = make(chan receivedRequest, 1)
		servers[name] = httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, request *stdhttp.Request) {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			received[name] <- receivedRequest{
				method:        request.Method,
				path:          request.URL.EscapedPath(),
				rawQuery:      request.URL.RawQuery,
				body:          string(body),
				authorization: request.Header.Get("Authorization"),
				correlationID: request.Header.Get(correlationHeader),
				host:          request.Host,
			}
			response.Header().Set("X-Legacy-Upstream", name)
			response.WriteHeader(stdhttp.StatusAccepted)
			_, _ = response.Write([]byte("from-" + name))
		}))
		defer servers[name].Close()
	}

	var logs bytes.Buffer
	registry := prometheus.NewRegistry()
	handler, err := NewHandler(Upstreams{
		Authz: servers["authz"].URL, Content: servers["content"].URL,
		Immersion: servers["immersion"].URL, Profile: servers["profile"].URL,
	}, stdhttp.DefaultTransport, time.Second, registry, slog.New(slog.NewJSONHandler(&logs, nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	facade := httptest.NewServer(handler)
	defer facade.Close()

	tests := []struct {
		name     string
		method   string
		path     string
		wantPath string
		body     string
	}{
		{name: "authz", method: stdhttp.MethodGet, path: "/authz/ping?detail=full", wantPath: "/ping"},
		{name: "content", method: stdhttp.MethodPost, path: "/content/pages/blog", wantPath: "/pages/blog", body: `{"title":"hello"}`},
		{name: "immersion", method: stdhttp.MethodPatch, path: "/immersion/logs/a%2Fb", wantPath: "/logs/a%2Fb", body: `{"amount":10}`},
		{name: "profile", method: stdhttp.MethodDelete, path: "/profile/users/old", wantPath: "/users/old"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := stdhttp.NewRequest(test.method, facade.URL+test.path, bytes.NewBufferString(test.body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			request.Host = "app.tadoku.test"
			request.Header.Set("Authorization", "Bearer oathkeeper-identity")
			request.Header.Set(correlationHeader, "request-"+test.name)

			response, err := facade.Client().Do(request)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if response.StatusCode != stdhttp.StatusAccepted {
				t.Errorf("got %v, want %v", response.StatusCode, stdhttp.StatusAccepted)
			}
			if string(body) != "from-"+test.name {
				t.Errorf("got %v, want %v", string(body), "from-"+test.name)
			}
			if response.Header.Get("X-Legacy-Upstream") != test.name {
				t.Errorf("got %v, want %v", response.Header.Get("X-Legacy-Upstream"), test.name)
			}
			if response.Header.Get(correlationHeader) != "request-"+test.name {
				t.Errorf("got %v, want %v", response.Header.Get(correlationHeader), "request-"+test.name)
			}

			var got receivedRequest
			select {
			case got = <-received[test.name]:
			case <-time.After(time.Second):
				t.Fatal("request did not reach expected upstream")
			}
			if got.method != test.method {
				t.Errorf("got %v, want %v", got.method, test.method)
			}
			if got.body != test.body {
				t.Errorf("got %v, want %v", got.body, test.body)
			}
			if got.authorization != "Bearer oathkeeper-identity" {
				t.Errorf("got %v, want %v", got.authorization, "Bearer oathkeeper-identity")
			}
			if got.correlationID != "request-"+test.name {
				t.Errorf("got %v, want %v", got.correlationID, "request-"+test.name)
			}
			if got.host != "app.tadoku.test" {
				t.Errorf("got %v, want %v", got.host, "app.tadoku.test")
			}
			if got.path != test.wantPath {
				t.Errorf("got %v, want %v", got.path, test.wantPath)
			}
			if test.name == "authz" {
				if got.rawQuery != "detail=full" {
					t.Errorf("got %v, want %v", got.rawQuery, "detail=full")
				}
			}
		})
	}

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	metricNames := make([]string, 0, len(metricFamilies))
	for _, family := range metricFamilies {
		metricNames = append(metricNames, family.GetName())
	}
	if !slices.Contains(metricNames, "tadoku_api_proxy_request_duration_seconds") {
		t.Errorf("missing %v in %v", "tadoku_api_proxy_request_duration_seconds", metricNames)
	}
	for _, test := range tests {
		if !strings.Contains(logs.String(), `"correlation_id":"request-`+test.name+`"`) {
			t.Errorf("missing %v in %v", `"correlation_id":"request-`+test.name+`"`, logs.String())
		}
		if !strings.Contains(logs.String(), `"upstream":"`+test.name+`"`) {
			t.Errorf("missing %v in %v", `"upstream":"`+test.name+`"`, logs.String())
		}
	}
}

func TestHandlerGeneratesAndForwardsCorrelationID(t *testing.T) {
	receivedID := make(chan string, 1)
	upstream := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, request *stdhttp.Request) {
		receivedID <- request.Header.Get(correlationHeader)
		response.WriteHeader(stdhttp.StatusNoContent)
	}))
	defer upstream.Close()

	var logs bytes.Buffer
	handler, err := NewHandler(Upstreams{
		Authz: upstream.URL, Content: upstream.URL, Immersion: upstream.URL, Profile: upstream.URL,
	}, stdhttp.DefaultTransport, time.Second, prometheus.NewRegistry(), slog.New(slog.NewJSONHandler(&logs, nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/content/ping", nil))

	generatedID := response.Header().Get(correlationHeader)
	if len(generatedID) == 0 {
		t.Errorf("expected a nonempty value")
	}
	if got := <-receivedID; got != generatedID {
		t.Errorf("got %v, want %v", got, generatedID)
	}
	if !strings.Contains(logs.String(), `"correlation_id":"`+generatedID+`"`) {
		t.Errorf("missing %v in %v", `"correlation_id":"`+generatedID+`"`, logs.String())
	}
}

func TestHandlerReturnsBadGatewayWhenUpstreamIsUnavailable(t *testing.T) {
	upstream := httptest.NewServer(stdhttp.NotFoundHandler())
	url := upstream.URL
	upstream.Close()

	handler := newTestHandler(t, url, time.Second)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/authz/ping", nil))

	if response.Code != stdhttp.StatusBadGateway {
		t.Errorf("got %v, want %v", response.Code, stdhttp.StatusBadGateway)
	}
	if response.Body.String() != "Bad Gateway\n" {
		t.Errorf("got %v, want %v", response.Body.String(), "Bad Gateway\n")
	}
	if len(response.Header().Get(correlationHeader)) == 0 {
		t.Errorf("expected a nonempty value")
	}
}

func TestHandlerTimesOutAndCancelsUpstreamRequest(t *testing.T) {
	cancelled := make(chan struct{})
	transport := roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
		<-request.Context().Done()
		close(cancelled)
		return nil, request.Context().Err()
	})

	handler := newTestHandlerWithTransport(t, transport, 20*time.Millisecond)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodPost, "/immersion/logs", bytes.NewBufferString("body")))

	if response.Code != stdhttp.StatusGatewayTimeout {
		t.Errorf("got %v, want %v", response.Code, stdhttp.StatusGatewayTimeout)
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("upstream request was not cancelled")
	}
}

func TestHandlerHealthAndUnknownRoutes(t *testing.T) {
	handler := newTestHandler(t, "http://127.0.0.1:1", time.Second)

	for _, test := range []struct {
		path   string
		status int
		body   string
	}{
		{path: "/livez", status: stdhttp.StatusOK, body: "ok"},
		{path: "/readyz", status: stdhttp.StatusOK, body: `{"status":"ready","checks":[]}`},
		{path: "/internal/v1/ping", status: stdhttp.StatusNotFound, body: "404 page not found\n"},
		{path: "/api/internal/content/ping", status: stdhttp.StatusNotFound, body: "404 page not found\n"},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, test.path, nil))
		if response.Code != test.status {
			t.Errorf("got %v, want %v", response.Code, test.status)
		}
		if response.Body.String() != test.body {
			t.Errorf("got %v, want %v", response.Body.String(), test.body)
		}
	}
}

func TestNewHandlerRejectsInvalidConfiguration(t *testing.T) {
	valid := Upstreams{
		Authz: "http://authz", Content: "http://content",
		Immersion: "http://immersion", Profile: "http://profile",
	}

	tests := []struct {
		name      string
		upstreams Upstreams
		timeout   time.Duration
	}{
		{name: "missing upstream", upstreams: Upstreams{}, timeout: time.Second},
		{name: "upstream path", upstreams: Upstreams{Authz: "http://authz/base", Content: valid.Content, Immersion: valid.Immersion, Profile: valid.Profile}, timeout: time.Second},
		{name: "timeout", upstreams: valid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewHandler(test.upstreams, stdhttp.DefaultTransport, test.timeout, prometheus.NewRegistry(), slog.Default())
			if err == nil {
				t.Errorf("expected an error")
			}
		})
	}
}

func newTestHandler(t testing.TB, upstream string, timeout time.Duration) stdhttp.Handler {
	t.Helper()
	handler, err := NewHandler(Upstreams{
		Authz: upstream, Content: upstream, Immersion: upstream, Profile: upstream,
	}, stdhttp.DefaultTransport, timeout, prometheus.NewRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return handler
}

func newTestHandlerWithTransport(t *testing.T, transport stdhttp.RoundTripper, timeout time.Duration) stdhttp.Handler {
	t.Helper()
	handler, err := NewHandler(Upstreams{
		Authz: "http://authz", Content: "http://content", Immersion: "http://immersion", Profile: "http://profile",
	}, transport, timeout, prometheus.NewRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return handler
}

type roundTripFunc func(*stdhttp.Request) (*stdhttp.Response, error)

func (fn roundTripFunc) RoundTrip(request *stdhttp.Request) (*stdhttp.Response, error) {
	return fn(request)
}

func TestRequestCancellationIsPreserved(t *testing.T) {
	started := make(chan struct{})
	cancelled := make(chan struct{})
	transport := roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
		close(started)
		<-request.Context().Done()
		close(cancelled)
		return nil, request.Context().Err()
	})

	handler := newTestHandlerWithTransport(t, transport, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(stdhttp.MethodGet, "/profile/users", nil).WithContext(ctx)
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(httptest.NewRecorder(), request)
		close(done)
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not reach upstream")
	}
	cancel()

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("upstream request was not cancelled")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("proxy did not return after cancellation")
	}
}
