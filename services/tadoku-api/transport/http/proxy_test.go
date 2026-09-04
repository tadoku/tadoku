package http

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.NoError(t, err)
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
	require.NoError(t, err)
	facade := httptest.NewServer(handler)
	defer facade.Close()

	tests := []struct {
		name     string
		method   string
		path     string
		wantPath string
		body     string
	}{
		{name: "authz", method: stdhttp.MethodGet, path: "/api/internal/authz/ping?detail=full", wantPath: "/ping"},
		{name: "content", method: stdhttp.MethodPost, path: "/api/internal/content/pages/blog", wantPath: "/pages/blog", body: `{"title":"hello"}`},
		{name: "immersion", method: stdhttp.MethodPatch, path: "/api/internal/immersion/logs/a%2Fb", wantPath: "/logs/a%2Fb", body: `{"amount":10}`},
		{name: "profile", method: stdhttp.MethodDelete, path: "/api/internal/profile/users/old", wantPath: "/users/old"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := stdhttp.NewRequest(test.method, facade.URL+test.path, bytes.NewBufferString(test.body))
			require.NoError(t, err)
			request.Host = "app.tadoku.test"
			request.Header.Set("Authorization", "Bearer oathkeeper-identity")
			request.Header.Set(correlationHeader, "request-"+test.name)

			response, err := facade.Client().Do(request)
			require.NoError(t, err)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			assert.Equal(t, stdhttp.StatusAccepted, response.StatusCode)
			assert.Equal(t, "from-"+test.name, string(body))
			assert.Equal(t, test.name, response.Header.Get("X-Legacy-Upstream"))
			assert.Equal(t, "request-"+test.name, response.Header.Get(correlationHeader))

			var got receivedRequest
			select {
			case got = <-received[test.name]:
			case <-time.After(time.Second):
				require.Fail(t, "request did not reach expected upstream")
			}
			assert.Equal(t, test.method, got.method)
			assert.Equal(t, test.body, got.body)
			assert.Equal(t, "Bearer oathkeeper-identity", got.authorization)
			assert.Equal(t, "request-"+test.name, got.correlationID)
			assert.Equal(t, "app.tadoku.test", got.host)
			assert.Equal(t, test.wantPath, got.path)
			if test.name == "authz" {
				assert.Equal(t, "detail=full", got.rawQuery)
			}
		})
	}

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)
	metricNames := make([]string, 0, len(metricFamilies))
	for _, family := range metricFamilies {
		metricNames = append(metricNames, family.GetName())
	}
	assert.Contains(t, metricNames, "tadoku_api_proxy_request_duration_seconds")
	for _, test := range tests {
		assert.Contains(t, logs.String(), `"correlation_id":"request-`+test.name+`"`)
		assert.Contains(t, logs.String(), `"upstream":"`+test.name+`"`)
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
	require.NoError(t, err)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/api/internal/content/ping", nil))

	generatedID := response.Header().Get(correlationHeader)
	assert.NotEmpty(t, generatedID)
	assert.Equal(t, generatedID, <-receivedID)
	assert.Contains(t, logs.String(), `"correlation_id":"`+generatedID+`"`)
}

func TestHandlerReturnsBadGatewayWhenUpstreamIsUnavailable(t *testing.T) {
	upstream := httptest.NewServer(stdhttp.NotFoundHandler())
	url := upstream.URL
	upstream.Close()

	handler := newTestHandler(t, url, time.Second)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/api/internal/authz/ping", nil))

	assert.Equal(t, stdhttp.StatusBadGateway, response.Code)
	assert.Equal(t, "Bad Gateway\n", response.Body.String())
	assert.NotEmpty(t, response.Header().Get(correlationHeader))
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
	handler.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodPost, "/api/internal/immersion/logs", bytes.NewBufferString("body")))

	assert.Equal(t, stdhttp.StatusGatewayTimeout, response.Code)
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		require.Fail(t, "upstream request was not cancelled")
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
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, test.path, nil))
		assert.Equal(t, test.status, response.Code)
		assert.Equal(t, test.body, response.Body.String())
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
			assert.Error(t, err)
		})
	}
}

func newTestHandler(t *testing.T, upstream string, timeout time.Duration) stdhttp.Handler {
	t.Helper()
	handler, err := NewHandler(Upstreams{
		Authz: upstream, Content: upstream, Immersion: upstream, Profile: upstream,
	}, stdhttp.DefaultTransport, timeout, prometheus.NewRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	require.NoError(t, err)
	return handler
}

func newTestHandlerWithTransport(t *testing.T, transport stdhttp.RoundTripper, timeout time.Duration) stdhttp.Handler {
	t.Helper()
	handler, err := NewHandler(Upstreams{
		Authz: "http://authz", Content: "http://content", Immersion: "http://immersion", Profile: "http://profile",
	}, transport, timeout, prometheus.NewRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	require.NoError(t, err)
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
	request := httptest.NewRequest(stdhttp.MethodGet, "/api/internal/profile/users", nil).WithContext(ctx)
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(httptest.NewRecorder(), request)
		close(done)
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		require.Fail(t, "request did not reach upstream")
	}
	cancel()

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		require.Fail(t, "upstream request was not cancelled")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		require.Fail(t, "proxy did not return after cancellation")
	}
}
