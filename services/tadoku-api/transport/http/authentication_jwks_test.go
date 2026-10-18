package http

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type writerFunc func([]byte) (int, error)

func (write writerFunc) Write(p []byte) (int, error) {
	return write(p)
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

			_, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, 24*time.Hour, "", tenant.Deployment{}, slog.Default())
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

	_, err := NewJWTAuthentication(t.Context(), server.URL, 20*time.Millisecond, 24*time.Hour, "", tenant.Deployment{}, slog.Default())
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
		_, err := NewJWTAuthentication(lifetime, server.URL, time.Minute, 24*time.Hour, "", tenant.Deployment{}, slog.Default())
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

	authenticate, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, 24*time.Hour, "", tenant.Deployment{}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	jwks.Store(fmt.Sprintf(`{"keys":[%s,%s]}`, oldJWK, newJWK))

	subject := "33333333-3333-4333-8333-333333333333"
	issuedAt := jwt.TimeFunc().Add(-time.Minute).UTC().Truncate(time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &userClaims{
		Tenant: "tadoku/prod",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
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
		if user == nil || user.Subject != subject || !user.CreatedAt.Equal(issuedAt) {
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

	authenticate, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, 24*time.Hour, "", tenant.Deployment{}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	issuedAt := jwt.TimeFunc().Add(-time.Minute).UTC().Truncate(time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &userClaims{
		Tenant: "tadoku/prod",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "44444444-4444-4444-8444-444444444444",
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
	authenticate, err := NewJWTAuthentication(lifetime, server.URL, time.Minute, 24*time.Hour, "", tenant.Deployment{}, logger)
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	issuedAt := jwt.TimeFunc().Add(-time.Minute).UTC().Truncate(time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &userClaims{
		Tenant: "tadoku/prod",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "55555555-5555-4555-8555-555555555555",
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
