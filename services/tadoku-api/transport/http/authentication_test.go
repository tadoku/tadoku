package http

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

func TestAuthenticationRequiresConfiguration(t *testing.T) {
	if _, err := NewJWTAuthentication(nil, "http://jwks.test", time.Second, slog.Default()); err == nil {
		t.Error("missing lifetime context accepted")
	}

	for _, test := range []struct {
		name    string
		url     string
		timeout time.Duration
	}{
		{name: "missing URL", timeout: time.Second},
		{name: "missing timeout", url: "http://jwks.test"},
		{name: "negative timeout", url: "http://jwks.test", timeout: -time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewJWTAuthentication(t.Context(), test.url, test.timeout, slog.Default()); err == nil {
				t.Error("invalid authentication configuration accepted")
			}
		})
	}

	passthrough := func(next stdhttp.Handler) stdhttp.Handler { return next }
	_, err := NewHandler(app.New(nil, nil, nil), func(context.Context) error { return nil }, time.Second, prometheus.NewRegistry(), slog.Default(), nil, passthrough)
	if err == nil {
		t.Error("router accepted missing authentication middleware")
	}
	_, err = NewHandler(app.New(nil, nil, nil), func(context.Context) error { return nil }, time.Second, prometheus.NewRegistry(), slog.Default(), passthrough, nil)
	if err == nil {
		t.Error("router accepted missing banned-user middleware")
	}
}

func TestNewApplicationRoutesInheritSharedMiddleware(t *testing.T) {
	type authenticationContextKey struct{}

	var authenticationCompositions, banCompositions, authenticated, checkedBans int
	authenticatedRequest := authenticationContextKey{}
	authenticate := func(next stdhttp.Handler) stdhttp.Handler {
		authenticationCompositions++
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			authenticated++
			if _, ok := r.Context().Deadline(); !ok {
				t.Error("authentication did not inherit the request timeout")
			}
			if r.URL.Path == "/test/auth-failure" {
				w.WriteHeader(stdhttp.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authenticatedRequest, true)))
		})
	}
	rejectBanned := func(next stdhttp.Handler) stdhttp.Handler {
		banCompositions++
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			checkedBans++
			if authenticated, _ := r.Context().Value(authenticatedRequest).(bool); !authenticated {
				t.Error("ban check ran before authentication")
			}
			if r.URL.Path != "/test/a/b" {
				w.WriteHeader(stdhttp.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	router, err := NewHandler(app.New(nil, nil, nil), func(context.Context) error { return nil }, time.Second, prometheus.NewRegistry(), slog.Default(), authenticate, rejectBanned)
	if err != nil {
		t.Fatal(err)
	}
	if authenticationCompositions != 1 || banCompositions != 1 {
		t.Fatalf("authentication compositions=%d ban compositions=%d, want 1 each", authenticationCompositions, banCompositions)
	}
	for _, path := range []string{"/test/first", "/test/second"} {
		router.Handle("GET "+path, stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
			t.Error("application handler bypassed the shared ban check")
			w.WriteHeader(stdhttp.StatusNoContent)
		}))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, path, nil))
		if response.Code != stdhttp.StatusForbidden {
			t.Errorf("%s status=%d, want 403", path, response.Code)
		}
	}
	router.HandleFunc("GET /test/auth-failure", func(stdhttp.ResponseWriter, *stdhttp.Request) {
		t.Error("application handler bypassed failed authentication")
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/test/auth-failure", nil))
	if response.Code != stdhttp.StatusUnauthorized {
		t.Errorf("auth failure status=%d, want 401", response.Code)
	}
	router.HandleFunc("GET /test/{value}", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if value := r.PathValue("value"); value != "a/b" {
			t.Errorf("path value=%q, want %q", value, "a/b")
		}
		w.WriteHeader(stdhttp.StatusNoContent)
	})
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/test/a%2Fb", nil))
	if response.Code != stdhttp.StatusNoContent {
		t.Errorf("escaped path status=%d, want 204", response.Code)
	}
	if authenticationCompositions != 1 || banCompositions != 1 {
		t.Errorf("authentication compositions=%d ban compositions=%d, want 1 each", authenticationCompositions, banCompositions)
	}
	if authenticated != 4 || checkedBans != 3 {
		t.Errorf("authentication calls=%d ban checks=%d, want 4 and 3", authenticated, checkedBans)
	}
}

func TestAuthenticationRejectsFailedJWKSFetch(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{name: "provider unavailable", status: stdhttp.StatusServiceUnavailable, body: "unavailable"},
		{name: "invalid JSON", status: stdhttp.StatusOK, body: "not JSON"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			t.Cleanup(server.Close)

			_, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, slog.Default())
			if err == nil {
				t.Fatal("failed JWKS fetch accepted")
			}
			if test.status != stdhttp.StatusOK {
				if !errors.Is(err, keyfunc.ErrInvalidHTTPStatusCode) {
					t.Errorf("unexpected status error: %v", err)
				}
			} else {
				var syntaxError *json.SyntaxError
				if !errors.As(err, &syntaxError) {
					t.Errorf("unexpected JSON error: %v", err)
				}
			}
		})
	}
}

func TestAuthenticationBoundsJWKSFetch(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(_ stdhttp.ResponseWriter, r *stdhttp.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	_, err := NewJWTAuthentication(t.Context(), server.URL, 20*time.Millisecond, slog.Default())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("JWKS fetch error=%v, want deadline exceeded", err)
	}
}

func TestAuthenticationCancelsInitialJWKSFetchWithLifetime(t *testing.T) {
	fetchStarted := make(chan struct{})
	fetchCanceled := make(chan struct{})
	server := httptest.NewServer(stdhttp.HandlerFunc(func(_ stdhttp.ResponseWriter, r *stdhttp.Request) {
		close(fetchStarted)
		<-r.Context().Done()
		close(fetchCanceled)
	}))
	t.Cleanup(server.Close)

	lifetime, cancelLifetime := context.WithCancel(t.Context())
	result := make(chan error, 1)
	go func() {
		_, err := NewJWTAuthentication(lifetime, server.URL, time.Minute, slog.Default())
		result <- err
	}()

	select {
	case <-fetchStarted:
	case <-time.After(time.Second):
		t.Fatal("initial JWKS fetch did not start")
	}
	cancelLifetime()
	select {
	case <-fetchCanceled:
	case <-time.After(time.Second):
		t.Error("canceling authentication lifetime did not cancel the initial JWKS fetch")
	}
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("initial JWKS fetch error=%v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Error("authentication constructor kept waiting after its lifetime was canceled")
	}
}

func TestAuthenticationRefreshesUnknownSigningKey(t *testing.T) {
	oldPrivateKey := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	newSeed := make([]byte, ed25519.SeedSize)
	newSeed[0] = 1
	newPrivateKey := ed25519.NewKeyFromSeed(newSeed)
	oldJWK := fmt.Sprintf(`{"kty":"OKP","crv":"Ed25519","kid":"old-key","alg":"EdDSA","use":"sig","x":%q}`,
		base64.RawURLEncoding.EncodeToString(oldPrivateKey.Public().(ed25519.PublicKey)))
	newJWK := fmt.Sprintf(`{"kty":"OKP","crv":"Ed25519","kid":"new-key","alg":"EdDSA","use":"sig","x":%q}`,
		base64.RawURLEncoding.EncodeToString(newPrivateKey.Public().(ed25519.PublicKey)))

	var jwks atomic.Value
	jwks.Store(fmt.Sprintf(`{"keys":[%s]}`, oldJWK))
	var fetches atomic.Int32
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		fetches.Add(1)
		_, _ = w.Write([]byte(jwks.Load().(string)))
	}))
	t.Cleanup(server.Close)

	authenticate, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	jwks.Store(fmt.Sprintf(`{"keys":[%s,%s]}`, oldJWK, newJWK))

	issuedAt := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  "rotated-user",
			IssuedAt: jwt.NewNumericDate(issuedAt),
		},
	})
	token.Header["kid"] = "new-key"
	signed, err := token.SignedString(newPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	called := 0
	handler := authenticate(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		called++
		user := identity.FromContext(r.Context())
		if user == nil || user.Subject != "rotated-user" || !user.CreatedAt.Equal(issuedAt) {
			t.Errorf("unexpected rotated identity: %+v", user)
		}
		w.WriteHeader(stdhttp.StatusNoContent)
	}))

	request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+signed)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != stdhttp.StatusNoContent {
		t.Errorf("status=%d, want %d", response.Code, stdhttp.StatusNoContent)
	}
	if called != 1 {
		t.Errorf("downstream calls=%d, want 1", called)
	}
	if got := fetches.Load(); got != 2 {
		t.Errorf("JWKS fetched %d times, want startup and rotation fetches", got)
	}
}

func TestAuthenticationRateLimitsUnknownSigningKeyRefresh(t *testing.T) {
	var fetches atomic.Int32
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		fetches.Add(1)
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(server.Close)

	authenticate, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	privateKey := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  "unknown-user",
			IssuedAt: jwt.NewNumericDate(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
	})
	token.Header["kid"] = "unknown-key"
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	handler := authenticate(stdhttp.HandlerFunc(func(stdhttp.ResponseWriter, *stdhttp.Request) {
		t.Error("unknown signing key reached downstream handler")
	}))

	for range 10 {
		request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
		request.Header.Set("Authorization", "Bearer "+signed)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != stdhttp.StatusUnauthorized {
			t.Errorf("status=%d, want %d", response.Code, stdhttp.StatusUnauthorized)
		}
	}
	if got := fetches.Load(); got != 2 {
		t.Errorf("JWKS fetched %d times, want one startup and one rate-limited refresh", got)
	}
}

func TestAuthenticationCancelsRefreshWithLifetime(t *testing.T) {
	refreshStarted := make(chan struct{})
	refreshCanceled := make(chan struct{})
	var fetches atomic.Int32
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if fetches.Add(1) == 1 {
			_, _ = w.Write([]byte(`{"keys":[]}`))
			return
		}

		close(refreshStarted)
		<-r.Context().Done()
		close(refreshCanceled)
	}))
	t.Cleanup(server.Close)

	lifetime, cancelLifetime := context.WithCancel(t.Context())
	authenticate, err := NewJWTAuthentication(lifetime, server.URL, time.Second, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	privateKey := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  "canceled-user",
			IssuedAt: jwt.NewNumericDate(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
	})
	token.Header["kid"] = "unknown-key"
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	handler := authenticate(stdhttp.HandlerFunc(func(stdhttp.ResponseWriter, *stdhttp.Request) {
		t.Error("unknown signing key reached downstream handler")
	}))
	request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+signed)
	handlerDone := make(chan struct{})
	go func() {
		handler.ServeHTTP(httptest.NewRecorder(), request)
		close(handlerDone)
	}()

	select {
	case <-refreshStarted:
	case <-time.After(time.Second):
		t.Fatal("unknown signing key did not start a JWKS refresh")
	}
	cancelLifetime()
	select {
	case <-refreshCanceled:
	case <-time.After(time.Second):
		t.Error("canceling authentication lifetime did not cancel the JWKS provider request")
	}
	select {
	case <-handlerDone:
	case <-time.After(time.Second):
		t.Error("authentication kept waiting after its lifetime was canceled")
	}
}
