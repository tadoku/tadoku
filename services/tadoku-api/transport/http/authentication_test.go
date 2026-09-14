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
