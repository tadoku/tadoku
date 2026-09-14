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
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

func TestAuthenticationRequiresConfiguration(t *testing.T) {
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
			if _, err := NewJWTAuthentication(test.url, test.timeout); err == nil {
				t.Error("invalid authentication configuration accepted")
			}
		})
	}

	passthrough := func(next stdhttp.Handler) stdhttp.Handler { return next }
	_, err := NewHandler(app.New(nil), func(context.Context) error { return nil }, time.Second, slog.Default(), nil, passthrough)
	if err == nil {
		t.Error("router accepted missing authentication middleware")
	}
	_, err = NewHandler(app.New(nil), func(context.Context) error { return nil }, time.Second, slog.Default(), passthrough, nil)
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
	router, err := NewHandler(app.New(nil), func(context.Context) error { return nil }, time.Second, slog.Default(), authenticate, rejectBanned)
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

			_, err := NewJWTAuthentication(server.URL, time.Second)
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

	_, err := NewJWTAuthentication(server.URL, 20*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("JWKS fetch error=%v, want deadline exceeded", err)
	}
}

func TestAuthenticationUsesCachedKeys(t *testing.T) {
	// A deterministic synthetic key exercises provider caching independently of
	// the shared HTTP contract fixtures. It is never used outside this test.
	privateKey := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	publicKey := privateKey.Public().(ed25519.PublicKey)
	jwks := fmt.Sprintf(`{"keys":[{"kty":"OKP","crv":"Ed25519","kid":"cached-key","alg":"EdDSA","use":"sig","x":%q}]}`,
		base64.RawURLEncoding.EncodeToString(publicKey))
	var fetches atomic.Int32
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		fetches.Add(1)
		_, _ = w.Write([]byte(jwks))
	}))
	t.Cleanup(server.Close)

	authenticate, err := NewJWTAuthentication(server.URL, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	issuedAt := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, &userClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  "cached-user",
			IssuedAt: jwt.NewNumericDate(issuedAt),
		},
	})

	called := 0
	handler := authenticate(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		called++
		user := identity.FromContext(r.Context())
		if user == nil || user.Subject != "cached-user" || !user.CreatedAt.Equal(issuedAt) {
			t.Errorf("unexpected cached identity: %+v", user)
		}
		w.WriteHeader(stdhttp.StatusNoContent)
	}))

	for _, kid := range []string{"unknown-key", "cached-key"} {
		token.Header["kid"] = kid
		signed, err := token.SignedString(privateKey)
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
		request.Header.Set("Authorization", "Bearer "+signed)
		if kid == "cached-key" {
			server.Close()
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		want := stdhttp.StatusUnauthorized
		if kid == "cached-key" {
			want = stdhttp.StatusNoContent
		}
		if response.Code != want {
			t.Errorf("kid=%s status=%d, want %d", kid, response.Code, want)
		}
	}
	if called != 1 {
		t.Errorf("downstream calls=%d, want 1", called)
	}
	if got := fetches.Load(); got != 1 {
		t.Errorf("JWKS fetched %d times, want only startup fetch", got)
	}
}
