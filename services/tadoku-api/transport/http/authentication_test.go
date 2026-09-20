package http

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type writerFunc func([]byte) (int, error)

func (write writerFunc) Write(p []byte) (int, error) {
	return write(p)
}

func rsaJWK(kid string, publicKey *rsa.PublicKey) string {
	return fmt.Sprintf(`{"kty":"RSA","kid":%q,"alg":"RS256","use":"sig","n":%q,"e":"AQAB"}`,
		kid, base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()))
}

func TestAuthenticationRequiresConfiguration(t *testing.T) {
	for _, test := range []struct {
		name      string
		lifetime  context.Context
		url       string
		timeout   time.Duration
		maxAge    time.Duration
		logger    *slog.Logger
		wantError string
	}{
		{name: "missing lifetime", url: "http://jwks.test", timeout: time.Second, maxAge: 24 * time.Hour, logger: slog.Default(), wantError: "authentication lifetime context is required"},
		{name: "missing URL", lifetime: t.Context(), timeout: time.Second, maxAge: 24 * time.Hour, logger: slog.Default(), wantError: "JWKS URL is required"},
		{name: "missing timeout", lifetime: t.Context(), url: "http://jwks.test", maxAge: 24 * time.Hour, logger: slog.Default(), wantError: "JWKS fetch timeout must be positive"},
		{name: "negative timeout", lifetime: t.Context(), url: "http://jwks.test", timeout: -time.Second, maxAge: 24 * time.Hour, logger: slog.Default(), wantError: "JWKS fetch timeout must be positive"},
		{name: "missing maximum token age", lifetime: t.Context(), url: "http://jwks.test", timeout: time.Second, logger: slog.Default(), wantError: "maximum token age must be positive"},
		{name: "negative maximum token age", lifetime: t.Context(), url: "http://jwks.test", timeout: time.Second, maxAge: -time.Hour, logger: slog.Default(), wantError: "maximum token age must be positive"},
		{name: "missing logger", lifetime: t.Context(), url: "http://jwks.test", timeout: time.Second, maxAge: 24 * time.Hour, wantError: "logger is required"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewJWTAuthentication(test.lifetime, test.url, test.timeout, test.maxAge, "", test.logger)
			if err == nil || err.Error() != test.wantError {
				t.Errorf("error=%v, want %q", err, test.wantError)
			}
		})
	}

	passthrough := func(next stdhttp.Handler) stdhttp.Handler { return next }
	_, err := NewHandler(app.New(nil, nil, nil, nil, nil, nil, nil), func(context.Context) error { return nil }, time.Second, prometheus.NewRegistry(), slog.Default(), nil, passthrough)
	if err == nil {
		t.Error("router accepted missing authentication middleware")
	}
	_, err = NewHandler(app.New(nil, nil, nil, nil, nil, nil, nil), func(context.Context) error { return nil }, time.Second, prometheus.NewRegistry(), slog.Default(), passthrough, nil)
	if err == nil {
		t.Error("router accepted missing banned-user middleware")
	}
}

func TestAuthenticationTokenPolicy(t *testing.T) {
	oldTimeFunc := jwt.TimeFunc
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return now }
	defer func() { jwt.TimeFunc = oldTimeFunc }()

	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	edPublicKey, edPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	jwks := fmt.Sprintf(`{"keys":[%s,{"kty":"OKP","crv":"Ed25519","kid":"ed-key","alg":"EdDSA","use":"sig","x":%q}]}`,
		rsaJWK("rsa-key", &rsaPrivateKey.PublicKey), base64.RawURLEncoding.EncodeToString(edPublicKey))
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = w.Write([]byte(jwks))
	}))
	t.Cleanup(server.Close)

	authenticate, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, 24*time.Hour, "", slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	handler := authenticate(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))

	for _, test := range []struct {
		name       string
		issuedAt   time.Time
		issuer     string
		method     jwt.SigningMethod
		kid        string
		privateKey any
		want       int
	}{
		{name: "exact maximum age", issuedAt: now.Add(-24 * time.Hour), method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusNoContent},
		{name: "just beyond maximum age", issuedAt: now.Add(-24*time.Hour - time.Second), method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusUnauthorized},
		{name: "other issuer when unpinned", issuedAt: now.Add(-time.Minute), issuer: "https://other.example.test/", method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusNoContent},
		{name: "missing issuer when unpinned", issuedAt: now.Add(-time.Minute), method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusNoContent},
		{name: "non RS256 with matching key", issuedAt: now.Add(-time.Minute), method: jwt.SigningMethodEdDSA, kid: "ed-key", privateKey: edPrivateKey, want: stdhttp.StatusUnauthorized},
	} {
		t.Run(test.name, func(t *testing.T) {
			token := jwt.NewWithClaims(test.method, &userClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    test.issuer,
					Subject:   "policy-user",
					IssuedAt:  jwt.NewNumericDate(test.issuedAt),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				},
			})
			token.Header["kid"] = test.kid
			signed, err := token.SignedString(test.privateKey)
			if err != nil {
				t.Fatal(err)
			}

			request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
			request.Header.Set("Authorization", "Bearer "+signed)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
		})
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
	router, err := NewHandler(app.New(nil, nil, nil, nil, nil, nil, nil), func(context.Context) error { return nil }, time.Second, prometheus.NewRegistry(), slog.Default(), authenticate, rejectBanned)
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

			_, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, 24*time.Hour, "", slog.Default())
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

	_, err := NewJWTAuthentication(t.Context(), server.URL, 20*time.Millisecond, 24*time.Hour, "", slog.Default())
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
		_, err := NewJWTAuthentication(lifetime, server.URL, time.Minute, 24*time.Hour, "", slog.Default())
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
	oldPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	newPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	oldJWK := rsaJWK("old-key", &oldPrivateKey.PublicKey)
	newJWK := rsaJWK("new-key", &newPrivateKey.PublicKey)

	var jwks atomic.Value
	jwks.Store(fmt.Sprintf(`{"keys":[%s]}`, oldJWK))
	var fetches atomic.Int32
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		fetches.Add(1)
		_, _ = w.Write([]byte(jwks.Load().(string)))
	}))
	t.Cleanup(server.Close)

	authenticate, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, 24*time.Hour, "", slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	jwks.Store(fmt.Sprintf(`{"keys":[%s,%s]}`, oldJWK, newJWK))

	issuedAt := jwt.TimeFunc().Add(-time.Minute).UTC().Truncate(time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "rotated-user",
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(time.Hour)),
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

	authenticate, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, 24*time.Hour, "", slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	issuedAt := jwt.TimeFunc().Add(-time.Minute).UTC().Truncate(time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "unknown-user",
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(time.Hour)),
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
	releaseRefresh := make(chan struct{})
	refreshFailed := make(chan struct{})
	releaseRefreshFailure := make(chan struct{})
	var logRefreshFailure sync.Once
	var fetches atomic.Int32
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if fetches.Add(1) == 1 {
			_, _ = w.Write([]byte(`{"keys":[]}`))
			return
		}

		close(refreshStarted)
		select {
		case <-r.Context().Done():
			close(refreshCanceled)
		case <-releaseRefresh:
		}
	}))
	t.Cleanup(func() {
		close(releaseRefresh)
		server.Close()
	})
	logger := slog.New(slog.NewTextHandler(writerFunc(func(p []byte) (int, error) {
		logRefreshFailure.Do(func() {
			close(refreshFailed)
			<-releaseRefreshFailure
		})
		return len(p), nil
	}), nil))
	t.Cleanup(func() { close(releaseRefreshFailure) })

	lifetime, cancelLifetime := context.WithCancel(t.Context())
	authenticate, err := NewJWTAuthentication(lifetime, server.URL, time.Minute, 24*time.Hour, "", logger)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	issuedAt := jwt.TimeFunc().Add(-time.Minute).UTC().Truncate(time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "canceled-user",
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(time.Hour)),
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
	serve := func(ctx context.Context) <-chan struct{} {
		done := make(chan struct{})
		request := httptest.NewRequest(stdhttp.MethodGet, "/", nil).WithContext(ctx)
		request.Header.Set("Authorization", "Bearer "+signed)
		go func() {
			handler.ServeHTTP(httptest.NewRecorder(), request)
			close(done)
		}()
		return done
	}
	requestContext, cancelRequest := context.WithCancel(t.Context())
	handlerDone := serve(requestContext)

	select {
	case <-refreshStarted:
	case <-time.After(time.Second):
		t.Fatal("unknown signing key did not start a JWKS refresh")
	}
	cancelRequest()
	select {
	case <-handlerDone:
	case <-time.After(time.Second):
		t.Error("authentication kept waiting after its request was canceled")
	}

	cancelLifetime()
	select {
	case <-refreshCanceled:
	case <-time.After(time.Second):
		t.Error("canceling authentication lifetime did not cancel the JWKS provider request")
	}
	select {
	case <-refreshFailed:
	case <-time.After(time.Second):
		t.Fatal("authentication did not finish the canceled JWKS refresh")
	}

	stoppedRequestContext, cancelStoppedRequest := context.WithCancel(context.Background())
	t.Cleanup(cancelStoppedRequest)
	select {
	case <-serve(stoppedRequestContext):
	case <-time.After(time.Second):
		t.Error("authentication kept waiting after its lifetime ended")
	}
}
